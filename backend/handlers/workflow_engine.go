package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"oa-system/database"
	"oa-system/models"

	"github.com/kyodo-tech/orchid"
)

const workflowTimeout = 10 * time.Second

// instanceLockKey 执行相关业务逻辑
func instanceLockKey(bizType string, bizID int) string {
	return fmt.Sprintf("wf_lock:%s:%d", bizType, bizID)
}

type dagNodeApproval struct {
	Approvers []int `json:"approvers"`
}

// loadOrchidDefinitionByBiz 加载业务数据
func loadOrchidDefinitionByBiz(bizType string) (*models.OrchidWorkflowDefinition, error) {
	var def models.OrchidWorkflowDefinition
	if err := database.DB.Where("biz_type = ? AND is_active = ?", bizType, true).Order("id desc").First(&def).Error; err != nil {
		return nil, err
	}
	return &def, nil
}

type dagEdgeRaw struct {
	From      string `json:"from"`
	To        string `json:"to"`
	Condition string `json:"condition"`
	Label     string `json:"label"`
}

type dagRaw struct {
	Edges []dagEdgeRaw `json:"edges"`
}

// edgeConditionKey 执行相关业务逻辑
func edgeConditionKey(from, to string) string {
	return from + "->" + to
}

// isConditionExpr 校验输入或状态
func isConditionExpr(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	return strings.Contains(s, "=") ||
		strings.Contains(s, ":") ||
		strings.Contains(s, "!=") ||
		strings.Contains(s, ">") ||
		strings.Contains(s, "<") ||
		strings.Contains(s, "~") ||
		strings.Contains(s, "&&") ||
		strings.Contains(s, "||")
}

// loadOrchidWorkflowByBiz 加载业务数据
func loadOrchidWorkflowByBiz(bizType string) (*models.OrchidWorkflowDefinition, *orchid.Workflow, map[string]string, error) {
	def, err := loadOrchidDefinitionByBiz(bizType)
	if err != nil {
		return nil, nil, nil, err
	}
	wf := orchid.NewWorkflow(def.Name)
	if err := wf.Import([]byte(def.DagJSON)); err != nil {
		return nil, nil, nil, err
	}
	if len(wf.Nodes) == 0 {
		return nil, nil, nil, errors.New("workflow has no nodes")
	}

	var raw dagRaw
	condMap := map[string]string{}
	if err := json.Unmarshal([]byte(def.DagJSON), &raw); err == nil {
		for _, e := range raw.Edges {
			cond := strings.TrimSpace(e.Condition)
			if cond == "" && isConditionExpr(e.Label) {
				cond = strings.TrimSpace(e.Label)
			}
			if cond != "" {
				condMap[edgeConditionKey(e.From, e.To)] = cond
			}
		}
	}
	return def, wf, condMap, nil
}

// hasOrchidWorkflowForBiz 校验输入或状态
func hasOrchidWorkflowForBiz(bizType string) bool {
	_, wf, _, err := loadOrchidWorkflowByBiz(bizType)
	if err != nil {
		return false
	}
	return len(wf.Nodes) > 0
}

// firstNodeKeys 执行相关业务逻辑
func firstNodeKeys(wf *orchid.Workflow) []string {
	incoming := map[string]int{}
	for key := range wf.Nodes {
		incoming[key] = 0
	}
	for _, e := range wf.Edges {
		incoming[e.To]++
	}
	res := []string{}
	for key, in := range incoming {
		if in == 0 {
			res = append(res, key)
		}
	}
	sort.Strings(res)
	return res
}

// isStartLikeNode 校验输入或状态
func isStartLikeNode(nodeKey string, node *orchid.Node) bool {
	if node == nil {
		return false
	}
	if strings.EqualFold(nodeKey, "start") {
		return true
	}
	if strings.EqualFold(node.ActivityName, "start") {
		return true
	}
	if node.Config != nil {
		if uiRaw, ok := node.Config["_ui"]; ok {
			if ui, ok := uiRaw.(map[string]interface{}); ok {
				if t, ok := ui["type"].(string); ok && strings.EqualFold(t, "start") {
					return true
				}
			}
		}
	}
	return false
}

// isEndLikeNode 校验输入或状态
func isEndLikeNode(nodeKey string, node *orchid.Node) bool {
	if node == nil {
		return false
	}
	if strings.EqualFold(nodeKey, "end") {
		return true
	}
	if strings.EqualFold(node.ActivityName, "end") {
		return true
	}
	if node.Config != nil {
		if uiRaw, ok := node.Config["_ui"]; ok {
			if ui, ok := uiRaw.(map[string]interface{}); ok {
				if t, ok := ui["type"].(string); ok && strings.EqualFold(t, "end") {
					return true
				}
			}
		}
	}
	return false
}

// nextNodeKeys 执行相关业务逻辑
func nextNodeKeys(wf *orchid.Workflow, nodeKey string) []string {
	res := []string{}
	for _, e := range wf.Edges {
		if e.From == nodeKey {
			res = append(res, e.To)
		}
	}
	return res
}

// getEdgeCondition 获取数据
func getEdgeCondition(e *orchid.Edge) string {
	if e == nil {
		return ""
	}
	if e.Label != nil && *e.Label != "" {
		return *e.Label
	}
	return ""
}

// nextNodeKeysWithCondition 执行相关业务逻辑
func nextNodeKeysWithCondition(wf *orchid.Workflow, nodeKey string, bizContext map[string]interface{}, condMap map[string]string) []string {
	type edgeInfo struct {
		to        string
		condition string
	}
	var condEdges []edgeInfo
	var unconditional []edgeInfo

	for _, e := range wf.Edges {
		if e.From != nodeKey {
			continue
		}
		cond := strings.TrimSpace(condMap[edgeConditionKey(e.From, e.To)])
		if cond == "" && e.Label != nil && isConditionExpr(*e.Label) {
			cond = strings.TrimSpace(*e.Label)
		}
		if cond == "" {
			unconditional = append(unconditional, edgeInfo{to: e.To})
		} else {
			condEdges = append(condEdges, edgeInfo{to: e.To, condition: cond})
		}
	}

	log.Printf("[workflow.route] node=%s cond_edges=%d unconditional_edges=%d", nodeKey, len(condEdges), len(unconditional))

	if len(condEdges) == 0 {
		res := []string{}
		for _, e := range unconditional {
			res = append(res, e.to)
		}
		return res
	}

	matched := []string{}
	for _, e := range condEdges {
		if evalCondition(e.condition, bizContext) {
			matched = append(matched, e.to)
		}
	}
	if len(matched) > 0 {
		return matched
	}
	res := []string{}
	for _, e := range unconditional {
		res = append(res, e.to)
	}
	return res
}

// evalCondition 执行相关业务逻辑
func evalCondition(condition string, ctx map[string]interface{}) bool {
	if strings.TrimSpace(condition) == "" || ctx == nil {
		return true
	}
	condition = strings.TrimSpace(condition)

	// OR: a || b
	orParts := splitByToken(condition, "||")
	if len(orParts) > 1 {
		for _, p := range orParts {
			if evalCondition(p, ctx) {
				return true
			}
		}
		return false
	}

	// AND: a && b
	andParts := splitByToken(condition, "&&")
	if len(andParts) > 1 {
		for _, p := range andParts {
			if !evalCondition(p, ctx) {
				return false
			}
		}
		return true
	}

	return evalAtomicCondition(condition, ctx)
}

// splitByToken 执行相关业务逻辑
func splitByToken(s, token string) []string {
	parts := strings.Split(s, token)
	res := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			res = append(res, p)
		}
	}
	if len(res) == 0 {
		return []string{s}
	}
	return res
}

// trimQuote 执行相关业务逻辑
func trimQuote(s string) string {
	s = strings.TrimSpace(s)
	s = strings.Trim(s, "\"")
	s = strings.Trim(s, "'")
	return s
}

// toFloat 执行相关业务逻辑
func toFloat(v interface{}) (float64, bool) {
	s := strings.TrimSpace(fmt.Sprintf("%v", v))
	if s == "" {
		return 0, false
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, false
	}
	return f, true
}

// evalAtomicCondition 执行相关业务逻辑
func evalAtomicCondition(condition string, ctx map[string]interface{}) bool {
	condition = strings.TrimSpace(condition)
	ops := []string{"<=", ">=", "!=", "!~", "~", "=", ":", ">", "<"}
	for _, op := range ops {
		if idx := strings.Index(condition, op); idx > 0 {
			key := strings.TrimSpace(condition[:idx])
			rawVal := trimQuote(condition[idx+len(op):])
			ctxVal, ok := ctx[key]
			if !ok {
				return false
			}
			left := fmt.Sprintf("%v", ctxVal)
			switch op {
			case "=", ":":
				return strings.EqualFold(strings.TrimSpace(left), rawVal)
			case "!=":
				return !strings.EqualFold(strings.TrimSpace(left), rawVal)
			case "~":
				return strings.Contains(strings.ToLower(left), strings.ToLower(rawVal))
			case "!~":
				return !strings.Contains(strings.ToLower(left), strings.ToLower(rawVal))
			case ">", "<", ">=", "<=":
				lf, lok := toFloat(ctxVal)
				rf, rok := toFloat(rawVal)
				if !lok || !rok {
					return false
				}
				switch op {
				case ">":
					return lf > rf
				case "<":
					return lf < rf
				case ">=":
					return lf >= rf
				case "<=":
					return lf <= rf
				}
			}
		}
	}
	// 未识别语法：默认放行，避免错误配置导致流程阻断
	return true
}

// loadBizContext 加载业务数据
func loadBizContext(bizType string, bizID int) map[string]interface{} {
	ctx := map[string]interface{}{}

	tableByBiz := map[string]string{
		"employee":      "employees",
		"leave_request": "leave_requests",
		"event_booking": "event_bookings",
	}
	tableName, ok := tableByBiz[bizType]
	if !ok {
		return ctx
	}

	row := map[string]interface{}{}
	if err := database.DB.Table(tableName).Where("id = ?", bizID).Take(&row).Error; err == nil {
		for k, v := range row {
			// 兼容旧写法：直接字段名
			ctx[k] = v
			// 新增：支持表前缀，避免跨业务同名字段歧义
			ctx[tableName+"."+k] = v // 例如 employees.status
			ctx[bizType+"."+k] = v   // 例如 employee.status
		}
	}

	// 关联字段（员工流程）：支持 department.name / position.name 形式
	if bizType == "employee" {
		if deptID, ok := ctx["department_id"]; ok {
			var dept models.Department
			if database.DB.First(&dept, deptID).Error == nil {
				ctx["department.name"] = dept.Name
				ctx["employee.department.name"] = dept.Name
				ctx["employees.department.name"] = dept.Name
			}
		}
		if posID, ok := ctx["position_id"]; ok {
			var pos models.Position
			if database.DB.First(&pos, posID).Error == nil {
				ctx["position.name"] = pos.Name
				ctx["employee.position.name"] = pos.Name
				ctx["employees.position.name"] = pos.Name
			}
		}
	}
	return ctx
}

// resolveAssigneeUserIDs 执行相关业务逻辑
func resolveAssigneeUserIDs(node *orchid.Node) []int {
	userIDs := []int{}
	seen := map[int]bool{}
	if node == nil {
		return userIDs
	}

	if cfgApprovers, ok := node.Config["approvers"]; ok {
		tmp := []int{}
		b, _ := json.Marshal(cfgApprovers)
		_ = json.Unmarshal(b, &tmp)
		for _, uid := range tmp {
			if uid > 0 && !seen[uid] {
				seen[uid] = true
				userIDs = append(userIDs, uid)
			}
		}
	}

	if cfgPos, ok := node.Config["approver_position_ids"]; ok {
		posIDs := []int{}
		b, _ := json.Marshal(cfgPos)
		_ = json.Unmarshal(b, &posIDs)
		if len(posIDs) > 0 {
			var emps []models.Employee
			database.DB.Where("position_id IN ? AND status = 1", posIDs).Find(&emps)
			for _, e := range emps {
				if e.UserID > 0 && !seen[e.UserID] {
					seen[e.UserID] = true
					userIDs = append(userIDs, e.UserID)
				}
			}
		}
	}
	return userIDs
}

// createTasksForNode 创建数据
func createTasksForNode(insID int, wf *orchid.Workflow, nodeKey string) {
	node := wf.Nodes[nodeKey]
	if node == nil {
		return
	}
	if isStartLikeNode(nodeKey, node) || isEndLikeNode(nodeKey, node) {
		return
	}
	uids := resolveAssigneeUserIDs(node)
	if len(uids) == 0 {
		return
	}

	type result struct{ err error }
	ch := make(chan result, len(uids))
	var wg sync.WaitGroup

	for _, uid := range uids {
		wg.Add(1)
		go func(assigneeID int) {
			defer wg.Done()
			err := database.DB.Create(&models.OrchidWorkflowTask{
				InstanceID: insID,
				NodeKey:    nodeKey,
				AssigneeID: assigneeID,
				Status:     "open",
			}).Error
			ch <- result{err: err}
		}(uid)
	}

	wg.Wait()
	close(ch)
	for r := range ch {
		if r.err != nil {
			log.Printf("[createTasksForNode] insID=%d node=%s err=%v", insID, nodeKey, r.err)
		}
	}
}

// startOrchidInstance 启动处理流程
func startOrchidInstance(bizType string, bizID int, operator string) (*models.OrchidWorkflowInstance, error) {
	ctx, cancel := context.WithTimeout(context.Background(), workflowTimeout)
	defer cancel()

	// 分布式锁，防止同一业务重复提交
	release, err := database.AcquireLock(ctx, instanceLockKey(bizType, bizID))
	if err != nil {
		return nil, err
	}
	defer release()

	def, wf, condMap, err := loadOrchidWorkflowByBiz(bizType)
	if err != nil {
		return nil, err
	}

	roots := firstNodeKeys(wf)
	if len(roots) == 0 {
		return nil, errors.New("workflow has no root node")
	}

	startCandidates := []string{}
	for _, k := range roots {
		if isStartLikeNode(k, wf.Nodes[k]) {
			startCandidates = append(startCandidates, k)
		}
	}
	if len(startCandidates) == 0 {
		startCandidates = roots
	}

	bizCtx := loadBizContext(bizType, bizID)

	// 从 start 节点用条件路由取第一层业务节点
	// - 有条件边：只取条件匹配的（条件分支）
	// - 无条件边：取所有后继（真正并行）
	firstBizNodes := []string{}
	seenFirst := map[string]bool{}
	for _, s := range startCandidates {
		nexts := nextNodeKeysWithCondition(wf, s, bizCtx, condMap)
		for _, nk := range nexts {
			if !isStartLikeNode(nk, wf.Nodes[nk]) && !seenFirst[nk] {
				seenFirst[nk] = true
				firstBizNodes = append(firstBizNodes, nk)
			}
		}
	}

	// 兜底1：条件全不匹配（或 start 没有连线）→ 忽略条件，取 start 的所有直接后继中第一个非 start/end 业务节点
	if len(firstBizNodes) == 0 {
		for _, s := range startCandidates {
			for _, nk := range nextNodeKeys(wf, s) {
				node := wf.Nodes[nk]
				if !isStartLikeNode(nk, node) && !isEndLikeNode(nk, node) && !seenFirst[nk] {
					seenFirst[nk] = true
					firstBizNodes = append(firstBizNodes, nk)
				}
			}
		}
		// 只取第一个，避免把所有分支节点都塞进去
		if len(firstBizNodes) > 1 {
			sort.Strings(firstBizNodes)
			firstBizNodes = firstBizNodes[:1]
		}
	}

	// 兜底2：start 完全孤立，取所有非 start 根节点中第一个
	if len(firstBizNodes) == 0 {
		for _, rk := range roots {
			if !isStartLikeNode(rk, wf.Nodes[rk]) && !isEndLikeNode(rk, wf.Nodes[rk]) && !seenFirst[rk] {
				seenFirst[rk] = true
				firstBizNodes = append(firstBizNodes, rk)
				break
			}
		}
	}

	sort.Strings(firstBizNodes)
	currentNodes := firstBizNodes

	if len(currentNodes) == 0 {
		return nil, errors.New("workflow has no executable node")
	}

	ins := models.OrchidWorkflowInstance{
		DefinitionID: def.ID,
		BizType:      bizType,
		BizID:        bizID,
		Status:       "pending",
		CurrentNode:  strings.Join(currentNodes, ","),
		StartedBy:    operator,
		StartedAt:    time.Now(),
	}
	if err := database.DB.Create(&ins).Error; err != nil {
		return nil, err
	}

	for _, nk := range currentNodes {
		node := wf.Nodes[nk]
		if isEndLikeNode(nk, node) {
			_ = database.DB.Create(&models.OrchidWorkflowHistory{InstanceID: ins.ID, NodeKey: nk, Action: "pending", Operator: operator, Remark: "到达结束节点"}).Error
			continue
		}
		if isStartLikeNode(nk, node) {
			_ = database.DB.Create(&models.OrchidWorkflowHistory{InstanceID: ins.ID, NodeKey: nk, Action: "pending", Operator: operator, Remark: "进入开始节点"}).Error
			continue
		}
		uids := resolveAssigneeUserIDs(node)
		if len(uids) == 0 {
			_ = database.DB.Create(&models.OrchidWorkflowHistory{InstanceID: ins.ID, NodeKey: nk, Action: "pending", Operator: operator, Remark: "无审批人，节点待办为空"}).Error
			continue
		}
		for _, uid := range uids {
			_ = database.DB.Create(&models.OrchidWorkflowTask{InstanceID: ins.ID, NodeKey: nk, AssigneeID: uid, Status: "open"}).Error
		}
		_ = database.DB.Create(&models.OrchidWorkflowHistory{InstanceID: ins.ID, NodeKey: nk, Action: "pending", Operator: operator, Remark: "进入待办"}).Error
	}

	_ = database.DB.Create(&models.OrchidWorkflowHistory{
		InstanceID: ins.ID,
		NodeKey:    "submit",
		Action:     "submit",
		Operator:   operator,
		Remark:     "提交进入流程",
	}).Error

	// 若初始即到达结束节点（例如条件直接命中 end），则自动完结流程
	allEnd := len(currentNodes) > 0
	for _, nk := range currentNodes {
		if !isEndLikeNode(nk, wf.Nodes[nk]) {
			allEnd = false
			break
		}
	}
	if allEnd {
		now := time.Now()
		ins.Status = "approved"
		ins.FinishedAt = &now
		ins.CurrentNode = "end"
		_ = database.DB.Save(&ins).Error
	}

	return &ins, nil
}

// getInstanceByBiz 获取数据
func getInstanceByBiz(bizType string, bizID int) (*models.OrchidWorkflowInstance, error) {
	var ins models.OrchidWorkflowInstance
	if err := database.DB.Where("biz_type = ? AND biz_id = ?", bizType, bizID).Order("id desc").First(&ins).Error; err != nil {
		return nil, err
	}
	return &ins, nil
}

// closeNodeTasks 执行相关业务逻辑
func closeNodeTasks(insID int, nodeKey string, assigneeID int, status string) {
	q := database.DB.Model(&models.OrchidWorkflowTask{}).Where("instance_id = ? AND node_key = ? AND status = 'open'", insID, nodeKey)
	if assigneeID > 0 {
		q = q.Where("assignee_id = ?", assigneeID)
	}
	_ = q.Update("status", status).Error
}

// openTaskCount 执行相关业务逻辑
func openTaskCount(insID int, nodeKey string) int64 {
	var c int64
	database.DB.Model(&models.OrchidWorkflowTask{}).Where("instance_id = ? AND node_key = ? AND status = 'open'", insID, nodeKey).Count(&c)
	return c
}

// parseCurrentNodes 解析输入数据
func parseCurrentNodes(current string) []string {
	nodes := []string{}
	if strings.TrimSpace(current) == "" {
		return nodes
	}
	seen := map[string]bool{}
	for _, part := range strings.Split(current, ",") {
		p := strings.TrimSpace(part)
		if p == "" || seen[p] {
			continue
		}
		seen[p] = true
		nodes = append(nodes, p)
	}
	sort.Strings(nodes)
	return nodes
}

// approveOrRejectInstance 处理审批业务
func approveOrRejectInstance(bizType string, bizID int, operator, action, remark string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), workflowTimeout)
	defer cancel()

	// 分布式锁，防止并发审批同一实例
	release, err := database.AcquireLock(ctx, instanceLockKey(bizType, bizID))
	if err != nil {
		return "", err
	}
	defer release()

	ins, err := getInstanceByBiz(bizType, bizID)
	if err != nil {
		return "", err
	}
	if ins.Status != "pending" {
		return ins.Status, nil
	}
	def, wf, condMap, err := loadOrchidWorkflowByBiz(bizType)
	if err != nil {
		return "", err
	}
	_ = def

	assigneeID := 0
	if operator != "" {
		var u models.User
		if err := database.DB.Where("real_name = ? OR username = ?", operator, operator).First(&u).Error; err == nil {
			assigneeID = u.ID
		}
	}

	currentNodes := parseCurrentNodes(ins.CurrentNode)
	if len(currentNodes) == 0 {
		return "", errors.New("instance has no current node")
	}

	nodeKey := currentNodes[0]
	if assigneeID > 0 {
		var task models.OrchidWorkflowTask
		if err := database.DB.Where("instance_id = ? AND status = 'open' AND assignee_id = ?", ins.ID, assigneeID).Order("id asc").First(&task).Error; err == nil {
			nodeKey = task.NodeKey
		}
	}

	if action == "rejected" {
		closeNodeTasks(ins.ID, nodeKey, assigneeID, "done")
		now := time.Now()
		ins.Status = "rejected"
		ins.FinishedAt = &now
		if err := database.DB.Save(ins).Error; err != nil {
			return "", err
		}
		_ = database.DB.Create(&models.OrchidWorkflowHistory{InstanceID: ins.ID, NodeKey: nodeKey, Action: "rejected", Operator: operator, Remark: remark}).Error
		return "rejected", nil
	}

	closeNodeTasks(ins.ID, nodeKey, assigneeID, "done")
	if openTaskCount(ins.ID, nodeKey) > 0 {
		_ = database.DB.Create(&models.OrchidWorkflowHistory{InstanceID: ins.ID, NodeKey: nodeKey, Action: "approved_partial", Operator: operator, Remark: remark}).Error
		return "pending", nil
	}

	bizCtx := loadBizContext(ins.BizType, ins.BizID)
	reachableSet := map[string]bool{}
	for _, ck := range currentNodes {
		if ck != nodeKey {
			// 未处理的并行节点保持激活
			reachableSet[ck] = true
			continue
		}

		currentNode := wf.Nodes[ck]
		// 结束节点不再继续向后路由，避免 end -> end 造成流程不收敛
		if isEndLikeNode(ck, currentNode) {
			reachableSet[ck] = true
			continue
		}

		nexts := nextNodeKeysWithCondition(wf, ck, bizCtx, condMap)
		if len(nexts) == 0 {
			reachableSet[ck] = true
			continue
		}
		for _, nk := range nexts {
			reachableSet[nk] = true
		}
	}

	for nk := range reachableSet {
		if openTaskCount(ins.ID, nk) > 0 {
			delete(reachableSet, nk)
		}
	}

	nextActiveSet := map[string]bool{}
	for nk := range reachableSet {
		node := wf.Nodes[nk]
		if isEndLikeNode(nk, node) {
			continue
		}

		if openTaskCount(ins.ID, nk) > 0 {
			nextActiveSet[nk] = true
			continue
		}

		uids := resolveAssigneeUserIDs(node)
		if len(uids) == 0 {
			_ = database.DB.Create(&models.OrchidWorkflowHistory{InstanceID: ins.ID, NodeKey: nk, Action: "pending", Operator: operator, Remark: "无审批人，节点待办为空"}).Error
			continue
		}
		for _, uid := range uids {
			_ = database.DB.Create(&models.OrchidWorkflowTask{InstanceID: ins.ID, NodeKey: nk, AssigneeID: uid, Status: "open"}).Error
		}
		_ = database.DB.Create(&models.OrchidWorkflowHistory{InstanceID: ins.ID, NodeKey: nk, Action: "pending", Operator: operator, Remark: "进入下一节点"}).Error
		nextActiveSet[nk] = true
	}

	nextActive := make([]string, 0, len(nextActiveSet))
	for nk := range nextActiveSet {
		nextActive = append(nextActive, nk)
	}

	if len(nextActive) == 0 {
		now := time.Now()
		ins.Status = "approved"
		ins.FinishedAt = &now
		ins.CurrentNode = "end"
		if err := database.DB.Save(ins).Error; err != nil {
			return "", err
		}
		_ = database.DB.Create(&models.OrchidWorkflowHistory{InstanceID: ins.ID, NodeKey: nodeKey, Action: "approved", Operator: operator, Remark: remark}).Error
		_ = database.DB.Create(&models.OrchidWorkflowHistory{InstanceID: ins.ID, NodeKey: "end", Action: "pending", Operator: operator, Remark: "到达结束节点"}).Error
		return "approved", nil
	}

	sort.Strings(nextActive)
	ins.CurrentNode = strings.Join(nextActive, ",")
	if err := database.DB.Save(ins).Error; err != nil {
		return "", err
	}
	_ = database.DB.Create(&models.OrchidWorkflowHistory{InstanceID: ins.ID, NodeKey: nodeKey, Action: "approved", Operator: operator, Remark: remark}).Error
	return "pending", nil
}

// transferTaskForInstance 执行相关业务逻辑
func transferTaskForInstance(bizType string, bizID int, fromUserID, toUserID int, operator, remark string) error {
	ins, err := getInstanceByBiz(bizType, bizID)
	if err != nil {
		return err
	}
	var task models.OrchidWorkflowTask
	if err := database.DB.Where("instance_id = ? AND assignee_id = ? AND status = 'open'", ins.ID, fromUserID).Order("id asc").First(&task).Error; err != nil {
		return err
	}
	task.Status = "transferred"
	if err := database.DB.Save(&task).Error; err != nil {
		return err
	}
	if err := database.DB.Create(&models.OrchidWorkflowTask{InstanceID: ins.ID, NodeKey: task.NodeKey, AssigneeID: toUserID, Status: "open"}).Error; err != nil {
		return err
	}
	return database.DB.Create(&models.OrchidWorkflowHistory{InstanceID: ins.ID, NodeKey: task.NodeKey, Action: "transfer", Operator: operator, Remark: remark}).Error
}

// skipCurrentNodeForInstance 执行相关业务逻辑
func skipCurrentNodeForInstance(bizType string, bizID int, operator, remark string) error {
	ins, err := getInstanceByBiz(bizType, bizID)
	if err != nil {
		return err
	}

	nodeKey := ""
	if operator != "" {
		var u models.User
		if err := database.DB.Where("real_name = ? OR username = ?", operator, operator).First(&u).Error; err == nil {
			var task models.OrchidWorkflowTask
			if err := database.DB.Where("instance_id = ? AND assignee_id = ? AND status = 'open'", ins.ID, u.ID).Order("id asc").First(&task).Error; err == nil {
				nodeKey = task.NodeKey
			}
		}
	}
	if nodeKey == "" {
		nodes := parseCurrentNodes(ins.CurrentNode)
		if len(nodes) > 0 {
			nodeKey = nodes[0]
		}
	}
	if nodeKey == "" {
		return errors.New("instance has no current node")
	}

	_ = database.DB.Model(&models.OrchidWorkflowTask{}).Where("instance_id = ? AND node_key = ? AND status = 'open'", ins.ID, nodeKey).Update("status", "skipped").Error
	_, err = approveOrRejectInstance(bizType, bizID, operator, "approved", "skip: "+remark)
	return err
}

// getOpenTasksForBiz 获取数据
func getOpenTasksForBiz(bizType string, bizID int) []models.OrchidWorkflowTask {
	ins, err := getInstanceByBiz(bizType, bizID)
	if err != nil {
		return []models.OrchidWorkflowTask{}
	}
	var tasks []models.OrchidWorkflowTask
	database.DB.Where("instance_id = ? AND status = 'open'", ins.ID).Find(&tasks)
	return tasks
}

// buildBizWorkflowLogs 构建业务数据
func buildBizWorkflowLogs(bizType string, bizID int) string {
	ins, err := getInstanceByBiz(bizType, bizID)
	if err != nil {
		b, _ := json.Marshal([]flowLogEntry{})
		return string(b)
	}
	var hs []models.OrchidWorkflowHistory
	database.DB.Where("instance_id = ?", ins.ID).Order("id asc").Find(&hs)
	logs := make([]flowLogEntry, 0, len(hs))
	for _, h := range hs {
		logs = append(logs, flowLogEntry{
			Time:     h.CreatedAt.Format(timeLayout),
			Node:     h.NodeKey,
			Action:   h.Action,
			Operator: h.Operator,
			Remark:   h.Remark,
		})
	}
	b, _ := json.Marshal(logs)
	return string(b)
}

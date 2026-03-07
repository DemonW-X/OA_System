package handlers

import (
	"encoding/json"
	"net/http"
	"oa-system/database"
	"oa-system/models"

	"github.com/gin-gonic/gin"
)

type GenerateWorkflowRequest struct {
	DepartmentID int    `json:"department_id" binding:"required"`
	BizType      string `json:"biz_type" binding:"required"`
	Name         string `json:"name" binding:"required"`
}

type PositionWithEmployees struct {
	Position  models.Position
	Employees []models.Employee
}

func GeneratePositionWorkflow(c *gin.Context) {
	var req GenerateWorkflowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}

	// 获取部门下的职位，按sort_order排序
	var positions []models.Position
	err := database.DB.
		Joins("JOIN department_positions dp ON dp.position_id = positions.id").
		Where("dp.department_id = ?", req.DepartmentID).
		Order("positions.sort_order ASC").
		Find(&positions).Error
	if err != nil || len(positions) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "该部门下没有职位"})
		return
	}

	// 获取每个职位下的员工
	var positionsWithEmps []PositionWithEmployees
	for _, pos := range positions {
		var emps []models.Employee
		database.DB.Where("position_id = ? AND status = 1", pos.ID).Find(&emps)
		if len(emps) > 0 {
			positionsWithEmps = append(positionsWithEmps, PositionWithEmployees{
				Position:  pos,
				Employees: emps,
			})
		}
	}

	if len(positionsWithEmps) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "该部门职位下没有在职员工"})
		return
	}

	// 生成DAG JSON
	dag := generatePositionDAG(positionsWithEmps)
	dagJSON, _ := json.Marshal(dag)

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"name":        req.Name,
			"biz_type":    req.BizType,
			"dag_json":    string(dagJSON),
			"positions":   positionsWithEmps,
			"description": "按职位排序自动生成的审批流程",
		},
	})
}

func generatePositionDAG(positionsWithEmps []PositionWithEmployees) map[string]interface{} {
	nodes := make(map[string]interface{})
	edges := []map[string]interface{}{}

	// 起始节点
	nodes["start"] = map[string]interface{}{
		"type": "start",
		"config": map[string]interface{}{
			"_ui": map[string]interface{}{
				"name": "开始",
				"x":    100,
				"y":    100,
			},
		},
	}

	prevNode := "start"
	yPos := 100

	// 为每个职位创建审批节点
	for i, pw := range positionsWithEmps {
		nodeKey := "approve_" + string(rune(i+1))
		yPos += 150

		// 获取该职位下所有员工的user_id
		var approvers []int
		for _, emp := range pw.Employees {
			if emp.UserID > 0 {
				approvers = append(approvers, emp.UserID)
			}
		}

		nodes[nodeKey] = map[string]interface{}{
			"type": "approve",
			"config": map[string]interface{}{
				"approvers": approvers,
				"_ui": map[string]interface{}{
					"name": pw.Position.Name + "审批",
					"x":    100,
					"y":    yPos,
				},
			},
		}

		// 连接上一个节点
		edges = append(edges, map[string]interface{}{
			"from": prevNode,
			"to":   nodeKey,
		})

		prevNode = nodeKey
	}

	// 结束节点
	yPos += 150
	nodes["end"] = map[string]interface{}{
		"type": "end",
		"config": map[string]interface{}{
			"_ui": map[string]interface{}{
				"name": "结束",
				"x":    100,
				"y":    yPos,
			},
		},
	}

	edges = append(edges, map[string]interface{}{
		"from": prevNode,
		"to":   "end",
	})

	return map[string]interface{}{
		"nodes": nodes,
		"edges": edges,
	}
}

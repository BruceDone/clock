package handler

import (
	"github.com/labstack/echo/v4"

	"clock/internal/domain"
	"clock/internal/logger"
	"clock/internal/service"
)

// RelationHandler 关系处理器
type RelationHandler struct {
	relationService service.RelationService
}

// NewRelationHandler 创建关系处理器
func NewRelationHandler(relationService service.RelationService) *RelationHandler {
	return &RelationHandler{
		relationService: relationService,
	}
}

// GetRelations 获取关系图
func (h *RelationHandler) GetRelations(c echo.Context) error {
	cid, err := getQueryInt(c, "cid")
	if err != nil {
		return BadRequest(c, err.Error())
	}

	graph, err := h.relationService.GetGraph(cid)
	if err != nil {
		logger.Errorf("[GetRelations] failed: %v", err)
		return HandleError(c, err)
	}

	return OK(c, graph)
}

// AddRelation 添加关系
func (h *RelationHandler) AddRelation(c echo.Context) error {
	var relation domain.Relation
	if err := c.Bind(&relation); err != nil {
		return BadRequest(c, "invalid request body")
	}

	if err := h.relationService.Add(&relation); err != nil {
		logger.Errorf("[AddRelation] failed: %v", err)
		return HandleError(c, err)
	}

	return OK(c, relation.Rid)
}

// DeleteRelation 删除关系
func (h *RelationHandler) DeleteRelation(c echo.Context) error {
	rid, err := getPathInt(c, "rid")
	if err != nil {
		return BadRequest(c, err.Error())
	}

	if err := h.relationService.Delete(rid); err != nil {
		logger.Errorf("[DeleteRelation] failed: %v", err)
		return HandleError(c, err)
	}

	return OK(c, nil)
}

package service

import (
	"clock/internal/domain"
	"clock/internal/repository"
	"clock/pkg/util"
)

// relationService 关系服务实现
type relationService struct {
	relationRepo repository.RelationRepository
	taskRepo     repository.TaskRepository
}

// NewRelationService 创建关系服务
func NewRelationService(
	relationRepo repository.RelationRepository,
	taskRepo repository.TaskRepository,
) RelationService {
	return &relationService{
		relationRepo: relationRepo,
		taskRepo:     taskRepo,
	}
}

// GetGraph 获取关系图
func (s *relationService) GetGraph(cid int) (*domain.RelationGraph, error) {
	tasks, err := s.taskRepo.GetByCID(cid)
	if err != nil {
		return nil, err
	}

	if len(tasks) == 0 {
		return &domain.RelationGraph{
			Nodes: []domain.Node{},
			Links: []domain.Link{},
		}, nil
	}

	relations, err := s.relationRepo.GetByCID(cid)
	if err != nil {
		return nil, err
	}

	return s.makeGraph(tasks, relations), nil
}

// Add 添加关系
func (s *relationService) Add(relation *domain.Relation) error {
	return s.relationRepo.Save(relation)
}

// Delete 删除关系
func (s *relationService) Delete(rid int) error {
	return s.relationRepo.Delete(rid)
}

// CheckCircle 使用拓扑排序检测DAG是否存在环
func (s *relationService) CheckCircle(tasks []*domain.Task, relations []*domain.Relation) bool {
	if len(tasks) == 0 || len(relations) == 0 {
		return false
	}

	// 复制任务列表
	taskList := make([]*domain.Task, len(tasks))
	copy(taskList, tasks)

	// 复制关系列表
	relationList := make([]*domain.Relation, len(relations))
	copy(relationList, relations)

	for {
		if len(taskList) == 0 {
			break
		}

		var rootTids []int

		// 初始化入度
		inDegree := make(map[int]int)
		for _, task := range taskList {
			inDegree[task.Tid] = 0
		}

		// 计算入度
		for _, rel := range relationList {
			if _, ok := inDegree[rel.NextTid]; ok {
				inDegree[rel.NextTid]++
			}
		}

		// 筛选入度为0的节点
		for tid, degree := range inDegree {
			if degree == 0 {
				rootTids = append(rootTids, tid)
			}
		}

		// 存在环
		if len(rootTids) == 0 {
			return true
		}

		// 移除节点
		taskList = util.Filter(taskList, func(t *domain.Task) bool {
			return !util.ContainsInt(rootTids, t.Tid)
		})

		// 移除关系
		relationList = util.Filter(relationList, func(r *domain.Relation) bool {
			return !util.ContainsInt(rootTids, r.Tid)
		})
	}

	return false
}

// makeGraph 构建关系图
func (s *relationService) makeGraph(tasks []*domain.Task, relations []*domain.Relation) *domain.RelationGraph {
	nodes := make([]domain.Node, 0, len(tasks))
	links := make([]domain.Link, 0, len(relations))

	for _, task := range tasks {
		nodes = append(nodes, task.ToNode())
	}

	for _, relation := range relations {
		links = append(links, relation.ToLink())
	}

	return &domain.RelationGraph{
		Nodes: nodes,
		Links: links,
	}
}

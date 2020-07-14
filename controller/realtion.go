package controller

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"

	"clock/param"
	"clock/storage"
)

// 得到当前所有的关系图
func GetRelations(c echo.Context) (err error) {
	cid, err := getQueryInt(c, "cid")

	resp := param.ApiResponse{
		Code: 200,
		Msg:  "success",
		Data: nil,
	}

	if err != nil {
		resp.Msg = fmt.Sprintf("[get relations] error to get the task tid with: %v", err)
		logrus.Error(resp.Msg)

		return c.JSON(http.StatusBadRequest, resp)
	}

	relation, err := storage.GetRelations(cid)

	if err != nil {
		resp.Msg = fmt.Sprintf("[get relations] query db error: %v", err)
		logrus.Error(resp.Msg)

		return c.JSON(http.StatusBadRequest, resp)
	}
	resp.Data = relation

	return c.JSON(http.StatusOK, resp)
}

// 删除关系
func DeleteRelation(c echo.Context) error {
	rid, err := getPathInt(c, "rid")

	resp := param.ApiResponse{
		Code: 200,
		Msg:  "success",
		Data: nil,
	}

	if err != nil {
		resp.Msg = fmt.Sprintf("[delete relation] find empty tid %v", err)
		logrus.Error(resp.Msg)

		return c.JSON(http.StatusBadRequest, resp)
	}

	if err := storage.DeleteRelation(rid); err != nil {
		resp.Msg = fmt.Sprintf("[delete relation] error to delete task with:%v", err)
		logrus.Error(err)

		return c.JSON(http.StatusBadRequest, resp)
	}

	return c.JSON(http.StatusOK, resp)
}

// 添加关系
func AddRelation(c echo.Context) error {
	resp := param.ApiResponse{
		Code: 200,
		Msg:  "success",
		Data: nil,
	}

	r := storage.Relation{}
	if err := c.Bind(&r); err != nil {
		resp.Msg = fmt.Sprintf("[add realtion] invalidate param found: %v", err)
		logrus.Error(resp.Msg)

		return c.JSON(http.StatusBadRequest, resp)
	}

	var tasks []storage.Task
	var relations []storage.Relation

	storage.Db.Where("cid = ?", r.Cid).Find(&tasks)
	storage.Db.Where("cid = ?", r.Cid).Find(&relations)
	relations = append(relations, r)

	// 存在闭环, 删除当前关系
	if circle := storage.CheckCircle(tasks,relations); circle {
		resp.Msg = fmt.Sprintf("[add realtion] check circle failed wit task: %d to task %d", r.Tid, r.NextTid)
		logrus.Warn(resp.Msg)
		resp.Data = -1

		return c.JSON(http.StatusOK, resp)
	}

	if err := storage.Db.Save(&r).Error ; err != nil{
		resp.Msg = fmt.Sprintf("[add realtion] error to add relation: %v", err)
		logrus.Error(resp.Msg)

		return c.JSON(http.StatusBadRequest, resp)
	}

	resp.Data = r.Rid
	return c.JSON(http.StatusOK, resp)
}

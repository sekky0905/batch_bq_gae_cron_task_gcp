package api

import (
	"github.com/gin-gonic/gin"
	"github.com/SekiguchiKai/batch_bq_gae_cron_task_gcp/server/model"
	"github.com/SekiguchiKai/batch_bq_gae_cron_task_gcp/server/util"
	"errors"
	"net/http"
	"time"
	"github.com/SekiguchiKai/batch_bq_gae_cron_task_gcp/server/store"
	"context"
)

// リクエストで受け取ったUserをDatastoreに新たに格納する。
func createUser(c *gin.Context) {
	util.InfoLog(c.Request, "createUser is called")

	var param model.User
	if err := bindUserFromJson(c, &param); err != nil {
		util.RespondAndLog(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := validateParamsForUser(param); err != nil {
		util.RespondAndLog(c, http.StatusBadRequest, err.Error())
		return
	}


	u := model.NewUser(param)
	u.CreatedAt = time.Now().UTC()
	u.UpdatedAt = time.Now().UTC()

	util.InfoLog(c.Request, "u :%+v", u)


	err := store.RunInTransaction(c.Request, func(ctx context.Context) error {
		s := store.NewUserStoreWithContext(ctx)

		if exists, err := s.ExistsUser(u.ID); err != nil {
			util.RespondAndLog(c, http.StatusInternalServerError, err.Error())
			return err
		}else if exists {
			util.RespondAndLog(c, http.StatusBadRequest, util.CreateErrMessage(_NotUniqueErrMessage).Error())
			return util.CreateErrMessage(_NotUniqueErrMessage)
		}

		if err := s.PutUser(u); err != nil {
			util.RespondAndLog(c, http.StatusInternalServerError, err.Error())
			return err
		}

		return nil
	})


	if err != nil {
		return
	}

	c.JSON(http.StatusOK, nil)

}


// HTTPのリクエストボディのjsonデータUserに変換する。
func bindUserFromJson(c *gin.Context, dst *model.User) error {
	if err := c.BindJSON(dst); err != nil {
		return err
	}

	dst.ID = getUserID(c)
	return nil
}

// URIのIDを取得する。
func getUserID(c *gin.Context) string {
	return c.Param("id")
}


// 送信されて来たUserに必要なデータが存在するかどうかのバリデーションを行う。
func validateParamsForUser(u model.User) error {
	if u.UserName == "" {
		return util.CreateErrMessage("UserName", _RequiredErrMessage)
	}

	if u.MailAddress == "" {
		return util.CreateErrMessage("MailAddress", _RequiredErrMessage)
	}

	if u.Age < 0 {
		return errors.New("Age should be over 0 ")
	}

	if u.Gender == "" {
		return util.CreateErrMessage("Gender", _RequiredErrMessage)
	}

	if u.From == "" {
		return util.CreateErrMessage("From", _RequiredErrMessage)
	}

	return nil

}
package bll

import (
	"context"
	"time"

	"github.com/LyricTian/gin-admin/src/model"
	"github.com/LyricTian/gin-admin/src/schema"
	"github.com/LyricTian/gin-admin/src/util"
	"github.com/pkg/errors"
)

// Demo 示例程序
type Demo struct {
	DemoModel model.IDemo `inject:"IDemo"`
}

// QueryPage 查询分页数据
func (a *Demo) QueryPage(ctx context.Context, params schema.DemoQueryParam, pageIndex, pageSize uint) (int64, []*schema.DemoQueryResult, error) {
	return a.DemoModel.QueryPage(ctx, params, pageIndex, pageSize)
}

// Get 查询指定数据
func (a *Demo) Get(ctx context.Context, recordID string) (*schema.Demo, error) {
	item, err := a.DemoModel.Get(ctx, recordID)
	if err != nil {
		return nil, err
	} else if item == nil {
		return nil, util.ErrNotFound
	}

	return item, nil
}

// Create 创建数据
func (a *Demo) Create(ctx context.Context, item *schema.Demo) error {
	exists, err := a.DemoModel.CheckCode(ctx, item.Code)
	if err != nil {
		return err
	} else if exists {
		return errors.New("编号已经存在")
	}

	item.ID = 0
	item.RecordID = util.MustUUID()
	item.Created = time.Now().Unix()
	item.Deleted = 0
	return a.DemoModel.Create(ctx, item)
}

// Update 更新数据
func (a *Demo) Update(ctx context.Context, recordID string, item *schema.Demo) error {
	oldItem, err := a.DemoModel.Get(ctx, recordID)
	if err != nil {
		return err
	} else if oldItem == nil {
		return util.ErrNotFound
	} else if oldItem.Code != item.Code {
		exists, err := a.DemoModel.CheckCode(ctx, item.Code)
		if err != nil {
			return err
		} else if exists {
			return errors.New("编号已经存在")
		}
	}

	info := util.StructToMap(item)
	delete(info, "id")
	delete(info, "record_id")
	delete(info, "creator")
	delete(info, "created")
	delete(info, "updated")
	delete(info, "deleted")

	return a.DemoModel.Update(ctx, recordID, info)
}

// Delete 删除数据
func (a *Demo) Delete(ctx context.Context, recordID string) error {
	exists, err := a.DemoModel.Check(ctx, recordID)
	if err != nil {
		return err
	} else if !exists {
		return util.ErrNotFound
	}

	return a.DemoModel.Delete(ctx, recordID)
}

// UpdateStatus 更新状态
func (a *Demo) UpdateStatus(ctx context.Context, recordID string, status int) error {
	exists, err := a.DemoModel.Check(ctx, recordID)
	if err != nil {
		return err
	} else if !exists {
		return util.ErrNotFound
	}

	info := map[string]interface{}{
		"status": status,
	}

	return a.DemoModel.Update(ctx, recordID, info)
}

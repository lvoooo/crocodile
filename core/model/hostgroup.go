package model

import (
	"context"
	"github.com/labulaka521/crocodile/common/db"
	"github.com/labulaka521/crocodile/common/log"
	"github.com/labulaka521/crocodile/common/utils"
	"github.com/labulaka521/crocodile/core/utils/define"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"math/rand"
	"strings"
	"time"
)

// CreateHostgroup create hostgroup
func CreateHostgroup(ctx context.Context, hg *define.HostGroup) error {
	createsql := `INSERT INTO crocodile_hostgroup (id,name,remark,createByID,hostIDs,createTime,updateTime) VALUES(?,?,?,?,?,?,?)`
	conn, err := db.GetConn(ctx)
	if err != nil {
		return errors.Wrap(err, "db.Db.GetConn")
	}
	defer conn.Close()
	stmt, err := conn.PrepareContext(ctx, createsql)
	if err != nil {
		return errors.Wrap(err, "conn.PrepareContext")
	}
	defer stmt.Close()
	createTime := time.Now().Unix()
	_, err = stmt.ExecContext(ctx,
		hg.ID,
		hg.Name,
		hg.Remark,
		hg.CreateByUID,
		strings.Join(hg.HostsID, ","),
		createTime,
		createTime)
	if err != nil {
		return errors.Wrap(err, "stmt.ExecContext")
	}
	return nil
}

// ChangeHostGroup change hostgroup
func ChangeHostGroup(ctx context.Context, hg *define.HostGroup) error {
	changesql := `UPDATE crocodile_hostgroup SET hostIDs=?,remark=?,updateTime=? WHERE id=?`
	conn, err := db.GetConn(ctx)
	if err != nil {
		return errors.Wrap(err, "db.Db.GetConn")
	}
	defer conn.Close()
	stmt, err := conn.PrepareContext(ctx, changesql)
	if err != nil {
		return errors.Wrap(err, "conn.PrepareContext")
	}
	defer stmt.Close()
	_, err = stmt.ExecContext(ctx, strings.Join(hg.HostsID, ","), hg.Remark, time.Now().Unix(), hg.ID)
	if err != nil {
		return errors.Wrap(err, "stmt.ExecContext")
	}
	return nil
}


// DeleteHostGroup delete hostgroup
func DeleteHostGroup(ctx context.Context, id string) error {
	sqldelete := `DELETE FROM crocodile_hostgroup WHERE id=?`
	conn, err := db.GetConn(ctx)
	if err != nil {
		return errors.Wrap(err, "db.Db.GetConn")
	}
	defer conn.Close()
	stmt, err := conn.PrepareContext(ctx, sqldelete)
	if err != nil {
		return errors.Wrap(err, "conn.PrepareContext")
	}
	defer stmt.Close()
	_, err = stmt.ExecContext(ctx, id)
	if err != nil {
		return errors.Wrap(err, "stmt.ExecContext")
	}
	return nil
}

// getHostGroups return hostgroup by id or hostgroupname
func getHostGroups(ctx context.Context, id, hgname string) ([]define.HostGroup, error) {
	hgs := []define.HostGroup{}

	sqlget := `SELECT 
					hg.id,
					hg.name,
					hg.remark,
					hg.hostIDs,
					hg.createByID,
					u.name,
					hg.createTime,
					hg.updateTime 
				FROM 
					crocodile_hostgroup as hg,crocodile_user as u
				WHERE
					hg.createByID == u.id`
	args := []interface{}{}
	if id != "" {
		sqlget += " AND hg.id=?"
		args = append(args, id)
	}
	if hgname != "" {
		sqlget += " AND hg.name=?"
		args = append(args, hgname)
	}
	conn, err := db.GetConn(ctx)
	if err != nil {
		return hgs, errors.Wrap(err, "db.Db.GetConn")
	}
	defer conn.Close()
	stmt, err := conn.PrepareContext(ctx, sqlget)
	if err != nil {
		return hgs, errors.Wrap(err, "conn.PrepareContext")
	}
	defer stmt.Close()
	rows, err := stmt.QueryContext(ctx, args...)
	if err != nil {
		return hgs, errors.Wrap(err, "stmt.QueryContext")
	}
	for rows.Next() {
		var (
			hg                     define.HostGroup
			addrs                  string
			createTime, updateTime int64
		)
		err := rows.Scan(&hg.ID, &hg.Name, &hg.Remark,
			&addrs, &hg.CreateByUID, &hg.CreateBy, &createTime, &updateTime)
		if err != nil {
			log.Info("Scan result failed", zap.Error(err))
			continue
		}
		hg.HostsID = []string{}
		if addrs != "" {
			hg.HostsID = append(hg.HostsID, strings.Split(addrs, ",")...)

		}
		hg.CreateTime = utils.UnixToStr(createTime)
		hg.UpdateTime = utils.UnixToStr(updateTime)

		hgs = append(hgs, hg)
	}
	return hgs, nil
}

// GetHostGroups return all hostgroup
func GetHostGroups(ctx context.Context) ([]define.HostGroup, error) {
	return getHostGroups(ctx, "", "")
}

// GetHostGroupID return hostgroup by id
func GetHostGroupID(ctx context.Context, id string) (*define.HostGroup, error) {
	hostgroups, err := getHostGroups(ctx, id, "")
	if err != nil {
		return nil, err
	}
	if len(hostgroups) != 1 {
		return nil, errors.New("can not find hostgroup id: " + id)
	}
	return &hostgroups[0], nil
}

// GetHostGroupName return hostgroup by name
func GetHostGroupName(ctx context.Context, hg string) (*define.HostGroup, error) {
	hostgroups, err := getHostGroups(ctx, "", hg)
	if err != nil {
		return nil, err
	}

	if len(hostgroups) != 1 {
		return nil, errors.New("can not find hostgroup name: " + hg)
	}
	return &hostgroups[0], nil
}

// RandHostID return execute worker ip
func RandHostID(hg *define.HostGroup) (string, error) {
	if len(hg.HostsID) == 0 {
		return "", errors.New("Can not find worker host")
	}
	hostid := hg.HostsID[rand.Int()%len(hg.HostsID)]
	return hostid, nil
}

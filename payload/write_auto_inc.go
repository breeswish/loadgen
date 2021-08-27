package payload

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/spf13/cobra"

	"github.com/crazycs520/loadgen/cmd"
	"github.com/crazycs520/loadgen/config"
	"github.com/crazycs520/loadgen/data"
	"github.com/crazycs520/loadgen/util"
)

type WriteAutoIncSuite struct {
	cfg        *config.Config
	tableName  string
	createTime time.Time
	seconds    int64
	inc        int64
	*basicWriteSuite
}

func NewWriteAutoIncSuite(cfg *config.Config) cmd.CMDGenerater {
	suite := &WriteAutoIncSuite{cfg: cfg, inc: -1}
	basic := NewBasicWriteSuite(cfg, suite)
	suite.basicWriteSuite = basic
	return suite
}

func (c *WriteAutoIncSuite) Name() string {
	return writeAutoIncSuiteName
}

func (c *WriteAutoIncSuite) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "write-auto-inc",
		Short:        "payload for write auto increment",
		RunE:         c.RunE,
		SilenceUsage: true,
	}
	cmd.Flags().Int64VarP(&c.seconds, "time", "", 600, "total time run testing (seconds)")
	return cmd
}

func (c *WriteAutoIncSuite) RunE(cmd *cobra.Command, args []string) error {
	fmt.Println("thread:", c.cfg.Thread)
	return c.Run()
}

func (c *WriteAutoIncSuite) Run() error {
	err := c.prepare()
	if err != nil {
		return err
	}
	fmt.Printf("start to do write auto increment for %v seconds\n", c.seconds)

	deadline := time.Now().Add(time.Duration(c.seconds) * time.Second)
	// TODO: implement a function to summarize result
	ctx, _ := context.WithDeadline(context.Background(), deadline)

	var wg sync.WaitGroup
	for i := 0; i < c.cfg.Thread; i++ {
		wg.Add(1)
		go func() {
			db := util.GetSQLCli(c.cfg)
			defer func() {
				db.Close()
			}()

			err := c.insertDataLoop(ctx, db)
			if err != nil {
				fmt.Println(err.Error())
			}
			wg.Done()
		}()
	}
	wg.Wait()

	return err
}

// empty function to implement payload.WriteSuite interface
func (c *WriteAutoIncSuite) UpdateTableDef(_ *data.TableInfo) {
}

// prepare defines and creates table used for test
func (c *WriteAutoIncSuite) prepare() error {
	c.tableName = "t_" + strings.ReplaceAll(c.Name(), "-", "_")
	tblInfo, err := data.NewTableInfo(c.cfg.DBName, c.tableName, []data.ColumnDef{
		{
			Name:     "auto",
			Tp:       "bigint",
			Property: "primary key auto_increment",
		},
		{
			Name: "a",
			Tp:   "bigint",
		},
		{
			Name: "b",
			Tp:   "bigint",
		},
		{
			Name:         "c",
			Tp:           "timestamp(6)",
			DefaultValue: "current_timestamp(6)",
		},
	}, []data.IndexInfo{
		{
			Name:    "idx0",
			Tp:      data.UniqueIndex,
			Columns: []string{"a"},
		},
	})
	if err != nil {
		return err
	}
	c.tblInfo = tblInfo
	load := data.NewLoadDataSuite(c.cfg)
	return load.CreateTable(c.tblInfo, false)
}

// insertDataLoop writes data into database
func (c *WriteAutoIncSuite) insertDataLoop(ctx context.Context, db *sql.DB) error {

	stmt, err := db.Prepare(c.genPrepareSQL())
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		inc := atomic.AddInt64(&c.inc, 1)
		args := c.genPrepareArgs(inc)
		_, err := stmt.Exec(args...)
		if err != nil {
			return err
		}
	}
}

func (c *WriteAutoIncSuite) genPrepareSQL() string {
	return "insert into " + c.tblInfo.DBTableName() + " (a,b) values (?, ?);"
}

func (c *WriteAutoIncSuite) genPrepareArgs(inc int64) []interface{} {
	return []interface{}{inc, inc}
}

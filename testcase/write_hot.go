package testcase

import (
	"fmt"

	"github.com/crazycs520/load/cmd"
	"github.com/crazycs520/load/config"
	"github.com/crazycs520/load/data"
	"github.com/spf13/cobra"
)

type WriteHotSuite struct {
	cfg       *config.Config
	tableName string
	tblInfo   *data.TableInfo

	rows        int
	insertCount int64
}

func NewWriteHotSuite(cfg *config.Config) cmd.CMDGenerater {
	return &WriteHotSuite{
		cfg: cfg,
	}
}

func (c *WriteHotSuite) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "write-hot",
		Short:        "stress test for write hot, such as auto_increment, timestamp index",
		RunE:         c.RunE,
		SilenceUsage: true,
	}
	cmd.Flags().IntVarP(&c.rows, "rows", "", 10000000, "total insert rows")
	return cmd
}

func (c *WriteHotSuite) RunE(cmd *cobra.Command, args []string) error {
	return c.Run()
}

func (c *WriteHotSuite) prepare() error {
	c.tableName = "t_write_hot"
	tblInfo, err := data.NewTableInfo(c.cfg.DBName, c.tableName, []data.ColumnDef{
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
		{
			Name: "d",
			Tp:   "varchar(50)",
		},
	}, []data.IndexInfo{
		{
			Tp:      data.PrimaryKey,
			Columns: []string{"a"},
		},
		{
			Tp:      data.NormalIndex,
			Columns: []string{"c", "d"},
		},
	})
	if err != nil {
		return err
	}
	c.tblInfo = tblInfo
	load := data.NewLoadDataSuit(c.cfg)
	return load.CreateTable(tblInfo, true)
}

func (c *WriteHotSuite) Run() error {
	err := c.prepare()
	if err != nil {
		fmt.Println("prepare table meet error: ", err)
		return err
	}

	load := data.NewLoadDataSuit(c.cfg)
	err = load.LoadData(c.tblInfo, c.rows)
	if err != nil {
		fmt.Printf("insert data error: %v\n", err)
	}
	return nil
}
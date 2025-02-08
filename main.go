package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/pflag"
)

func main() {
	var (
		options  CollectOptions
		delete   bool
		loadFile string
		saveFile = "filecompact.state"
		debug    bool
	)

	pflag.ErrHelp = errors.New("")
	pflag.BoolVar(&delete, "delete", delete, "delete files")
	pflag.StringVar(&loadFile, "load", loadFile, "load collect database file")
	pflag.StringVar(&saveFile, "save", saveFile, "save collect database file")
	pflag.StringSliceVarP(&options.Sources, "source", "s", options.Sources, "source directories")
	pflag.StringSliceVarP(&options.Exclude, "exclude", "e", options.Exclude, "exclude files")
	pflag.BoolVarP(&debug, "debug", "d", debug, "debug mode")
	pflag.BoolVarP(&options.Strict, "strict", "S", options.Strict, "strict mode")
	pflag.Parse()

	if debug {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	slog.Info("collecting...")

	var c Collection

	var err error
	if loadFile != "" {
		err = LoadCollection(loadFile, &c)
	} else {
		options.Sources = append(options.Sources, pflag.Args()...)
		if len(options.Sources) == 0 {
			options.Sources = []string{"."}
		}
		c, err = CollectFiles(options)
	}

	if err != nil {
		slog.Error("整理文件发生错误", "error", err)
		return
	}

	if !delete {
		slog.Info("file collect completed")
		for sig, files := range c.Files {
			slog.Info(sig, "count", len(files))
			for _, file := range files {
				slog.Info(fmt.Sprintf("  - %s %s", file.CreatedTime.Format("2006-01-02 15:04:05"), file.Path))
			}
		}
	}

	slog.Info("收集完成", "count", c.TotalFiles, "size", HumanSize(c.TotalSize), "elapsed", c.Elapsed, "files", len(c.Files))

	if loadFile == "" && saveFile != "" {
		if err = c.Save(saveFile); err != nil {
			slog.Error("保存收集结果到数据库文件发生错误", "error", err)
			return
		}
		slog.Info("收集结果保存到数据库文件完成")
	}

	if !delete {
		slog.Info("全部完成")
		return
	}

	defer os.Remove(saveFile)

	slog.Info("删除重复文件...")

	delResult, err := c.AutoDelete()
	if err != nil {
		slog.Error("删除重复文件发生错误", "error", err)
		return
	}

	if len(delResult.Errors) > 0 {
		slog.Info("一些文件删除失败:")
		for path, err := range delResult.Errors {
			slog.Info("  - "+path, "error", err)
		}
	}

	slog.Info("删除完成", "count", delResult.Deleted, "size", HumanSize(delResult.DeletedSize), "elapsed", delResult.Elapsed)
	slog.Info("全部完成")
}

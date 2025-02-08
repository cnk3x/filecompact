package main

import (
	"cmp"
	"crypto/md5"
	"crypto/sha256"
	"fmt"
	"iter"
	"log/slog"
	"os"
	"slices"
	"strings"
	"sync"
	"time"

	"go.etcd.io/bbolt"
)

// 文件信息
type FileInfo struct {
	Name         string
	Path         string
	Size         int64
	IsDir        bool
	CreatedTime  time.Time
	ModifiedTime time.Time
	MD532K       string
	SHA256       string
}

// 收集选项
type CollectOptions struct {
	Sources []string // 源目录
	Exclude []string // 排除目录
	Strict  bool     // 严格模式
}

// 收集结果
type Collection struct {
	CollectOptions `json:",inline"`

	Files       map[string][]*FileInfo `json:"files,omitempty"`        // 重复文件组
	TotalFiles  int                    `json:"count,omitempty"`        // 所有文件数量
	TotalSize   int64                  `json:"total_size,omitempty"`   // 总大小
	DeleteSize  int64                  `json:"delete_size,omitempty"`  // 可删除大小
	DeleteCount int64                  `json:"delete_count,omitempty"` // 可删除数量
	Elapsed     time.Duration          `json:"elapsed,omitempty"`      // 耗时
}

// 收集并查重
//
//	先按文件大小分组，分组中文件数量小于1个的分组丢弃，
//	再文件的MD532K摘要分组，分组中文件数量小于1个的分组丢弃，
//	如果是严格模式，则再次按文件全内容SHA256摘要分组，分组中文件数量小于1个的分组丢弃。
//	最后得到的分组，就是有重复文件的文件组。
func CollectFiles(cOpts CollectOptions) (result Collection, err error) {
	var st = time.Now()

	var files []*FileInfo
	defer clear(files)

	slog.Debug("收集文件")
	for _, source := range cOpts.Sources {
		slog.Debug(fmt.Sprintf(" - %s", source))
	}

	// 收集所有文件
	for _, source := range cOpts.Sources {
		if err = ListFiles(source, &files, cOpts.Exclude); err != nil {
			return
		}
	}

	slog.Debug(fmt.Sprintf("收集完成，共 %d 个文件", len(files)))

	groupBySize := func(files []*FileInfo) iter.Seq2[int64, []*FileInfo] {
		return func(yield func(int64, []*FileInfo) bool) {
			m := make(map[int64][]*FileInfo)
			for _, f := range files {
				m[f.Size] = append(m[f.Size], f)
			}

			defer clear(m)

			for size, files := range m {
				if len(files) > 1 && !yield(size, files) {
					return
				}
			}
		}
	}

	// 按文件摘要分组的方法
	groupByHash := func(files []*FileInfo, fullContent bool) iter.Seq2[string, []*FileInfo] {
		return func(yield func(string, []*FileInfo) bool) {
			m := haspMapPool.Get().(map[string][]*FileInfo)

			defer haspMapPool.Put(m)
			defer clear(m)

			for _, file := range files {
				hash := FileHash(file.Path, Iif(fullContent, sha256.New, md5.New), fullContent)
				if hash == "" || hash == ErrHash {
					continue
				}

				if fullContent {
					file.SHA256 = hash
				} else {
					file.MD532K = hash
				}

				m[hash] = append(m[hash], file)
			}

			for k, files := range m {
				if len(files) > 1 && !yield(k, files) {
					return
				}
			}
		}
	}

	slog.Debug("对文件进行分组")
	result.Files = make(map[string][]*FileInfo)
	// 按文件大小分组
	slog.Debug("按文件大小分组")
	for _, files := range groupBySize(files) {
		// 按文件MD532K摘要分组
		slog.Debug("按文件MD532K摘要分组")
		for md5, files := range groupByHash(files, false) {
			if cOpts.Strict {
				// 按文件全内容SHA256摘要分组
				slog.Debug("按文件全内容SHA256摘要分组")
				for sha, files := range groupByHash(files, true) {
					result.Files[sha] = append(result.Files[sha], files...)
				}
			} else {
				result.Files[md5] = append(result.Files[md5], files...)
			}
		}
	}

	slog.Debug("按文件创建时间顺序、路径长度顺序， 路径名称顺序排序")
	for hash, files := range result.Files {
		// 按文件创建时间顺序、路径长度顺序， 路径名称顺序排序
		slices.SortFunc(result.Files[hash], func(a, b *FileInfo) int {
			if r := a.CreatedTime.Compare(b.CreatedTime); r != 0 {
				return r
			}
			if r := cmp.Compare(len(a.Path), len(b.Path)); r != 0 {
				return r
			}
			return strings.Compare(a.Path, b.Path)
		})

		// 计算可删除的文件数量和大小
		fsl := int64(len(result.Files[hash]) - 1)
		result.DeleteCount += fsl
		result.DeleteSize += fsl * files[0].Size
	}

	result.TotalSize = SumBy(files, func(item *FileInfo) int64 { return item.Size })
	result.TotalFiles = len(files)
	result.Elapsed = time.Since(st)
	return
}

// 删除结果
type DelResult struct {
	CollectOptions `json:",inline"`
	Deleted        int               // 删除的文件数
	DeletedSize    int64             // 删除的文件大小
	Errors         map[string]string // 错误
	Elapsed        time.Duration     // 耗时
}

// 自动删除重复文件
func (c *Collection) AutoDelete() (result *DelResult, err error) {
	st := time.Now()

	result = &DelResult{CollectOptions: c.CollectOptions}

	for _, files := range c.Files {
		for i, file := range files {
			if i > 0 {
				if e := os.Remove(file.Path); e != nil {
					result.Errors[file.Path] = e.Error()
				} else {
					result.Deleted++
					result.DeletedSize += file.Size
					slog.Debug("删除文件", "file", file.Path)
				}
			}
		}
	}

	result.Elapsed = time.Since(st)
	return
}

func (c *Collection) Save(dbFile string) (err error) {
	slog.Debug("保存收集结果到数据库文件", "file", dbFile)

	var db *bbolt.DB
	if db, err = bbolt.Open(dbFile, 0600, nil); err != nil {
		slog.Error("保存收集结果到数据库文件发生错误", "error", err)
		return
	}
	defer db.Close()

	err = db.Update(func(tx *bbolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("collect"))
		if err != nil {
			return err
		}

		bucket.Put([]byte("sources"), ToBytes(c.Sources))
		bucket.Put([]byte("exclude"), ToBytes(c.Exclude))
		bucket.Put([]byte("total_files"), ToBytes(c.TotalFiles))
		bucket.Put([]byte("total_size"), ToBytes(c.TotalSize))
		bucket.Put([]byte("elapsed"), ToBytes(c.Elapsed))
		bucket.Put([]byte("delete_count"), ToBytes(c.DeleteCount))
		bucket.Put([]byte("delete_size"), ToBytes(c.DeleteSize))

		if groupsBucket := May(bucket.CreateBucketIfNotExists([]byte("files"))); groupsBucket != nil {
			for sig, files := range c.Files {
				groupsBucket.Put([]byte(sig), ToBytes(files))
			}
		}

		return nil
	})

	if err != nil {
		slog.Error("保存收集结果到数据库文件发生错误", "error", err)
		return
	}

	slog.Debug("收集结果保存到数据库文件完成", "file", dbFile)
	return

}

func LoadCollection(dbFile string, c *Collection) (err error) {
	slog.Debug("从数据库文件加载收集结果", "file", dbFile)

	var db *bbolt.DB
	if db, err = bbolt.Open(dbFile, 0600, nil); err != nil {
		slog.Error("从数据库文件加载收集结果发生错误", "error", err)
		return
	}
	defer db.Close()

	err = db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("collect"))
		if bucket == nil {
			return fmt.Errorf("bucket not found")
		}

		FromBytes(bucket.Get([]byte("sources")), &c.Sources)
		FromBytes(bucket.Get([]byte("exclude")), &c.Exclude)
		FromBytes(bucket.Get([]byte("total_files")), &c.TotalFiles)
		FromBytes(bucket.Get([]byte("total_size")), &c.TotalSize)
		FromBytes(bucket.Get([]byte("elapsed")), &c.Elapsed)
		FromBytes(bucket.Get([]byte("delete_count")), &c.DeleteCount)
		FromBytes(bucket.Get([]byte("delete_size")), &c.DeleteSize)

		c.Files = make(map[string][]*FileInfo)

		if groupsBucket := bucket.Bucket([]byte("files")); groupsBucket != nil {
			return groupsBucket.ForEach(func(k, v []byte) error {
				var files []*FileInfo
				FromBytes(v, &files)
				c.Files[string(k)] = files
				return nil
			})
		}
		return nil
	})

	if err != nil {
		slog.Error("从数据库文件加载收集结果发生错误", "error", err)
		return
	}

	slog.Debug("从数据库文件加载收集结果完成")

	return
}

var haspMapPool = sync.Pool{New: func() any { return make(map[string][]*FileInfo, 10) }}

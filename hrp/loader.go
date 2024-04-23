package hrp

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

func LoadTestCases(iTestCases ...ITestCase) ([]*TestCase, error) {
	testCases := make([]*TestCase, 0)

	for _, iTestCase := range iTestCases {
		if _, ok := iTestCase.(*TestCase); ok {
			testcase, err := iTestCase.ToTestCase()
			if err != nil {
				log.Error().Err(err).Msg("failed to convert ITestCase interface to TestCase struct")
				return nil, err
			}
			testCases = append(testCases, testcase)
			continue
		}

		// iTestCase should be a TestCasePath, file path or folder path
		tcPath, ok := iTestCase.(*TestCasePath)
		if !ok {
			return nil, errors.New("invalid iTestCase type")
		}

		casePath := tcPath.GetPath()
		err := fs.WalkDir(os.DirFS(casePath), ".", func(path string, dir fs.DirEntry, e error) error {
			if dir == nil {
				// casePath is a file other than a dir
				path = casePath
			} else if dir.IsDir() && path != "." && strings.HasPrefix(path, ".") {
				// skip hidden folders
				return fs.SkipDir
			} else {
				// casePath is a dir
				path = filepath.Join(casePath, path)
			}

			// ignore non-testcase files
			ext := filepath.Ext(path)
			if ext != ".yml" && ext != ".yaml" && ext != ".json" {
				return nil
			}

			// filtered testcases
			testCasePath := TestCasePath(path)
			tc, err := testCasePath.ToTestCase()
			if err != nil {
				log.Warn().Err(err).Str("path", path).Msg("load testcase failed")
				return err
			}
			testCases = append(testCases, tc)
			return nil
		})
		if err != nil {
			return nil, errors.Wrap(err, "read dir failed")
		}
	}

	log.Info().Int("count", len(testCases)).Msg("load testcases successfully")
	return testCases, nil
}

// 重构：按路径加载测试脚本，根据文件夹形式分配
func LoadTestCasesPaths(iTestCases ...ITestCase) ([]string, map[string][]string, error) {
	var dirMap = make(map[string][]string, 0)
	var dirList = make([]string, 0)

	for _, iTestCase := range iTestCases {
		// iTestCase should be a TestCasePath, file path or folder path
		tcPath, ok := iTestCase.(*TestCasePath)
		if !ok {
			return nil, nil, errors.New("invalid iTestCase type")
		}

		casePath := tcPath.GetPath()
		var paths = []string{}
		err := fs.WalkDir(os.DirFS(casePath), ".", func(path string, dir fs.DirEntry, e error) error {
			if dir == nil {
				// casePath is a file other than a dir
				path = casePath
			} else if dir.IsDir() && path != "." && strings.HasPrefix(path, ".") {
				// skip hidden folders
				return fs.SkipDir
			} else {
				// casePath is a dir
				path = filepath.Join(casePath, path)
			}

			// ignore non-testcase files
			ext := filepath.Ext(path)
			if ext != ".yml" && ext != ".yaml" && ext != ".json" {
				return nil
			}

			// filtered testcases
			paths = append(paths, path)

			return nil
		})
		if err != nil {
			return nil, nil, errors.Wrap(err, "read dir failed")
		}

		for _, path := range paths {
			dirName := filepath.Dir(path)
			if _, ok := dirMap[dirName]; !ok {
				dirList = append(dirList, dirName)
			}
			dirMap[dirName] = append(dirMap[dirName], path)
		}
	}

	return dirList, dirMap, nil
}

// 重构：分割组织好的文件夹测试脚本，用于协程执行时，平均分配测试任务
func splitDirList(dirList []string, parts int) [][]string {
	result := make([][]string, parts)

	if parts <= 0 {
		result = append(result, dirList)
		return result
	}

	if parts >= len(dirList) {
		// 如果要求分成的份数超过数组的长度，将每个元素作为一个部分
		for i := range dirList {
			result = append(result, []string{dirList[i]})
		}
		return result
	}

	partSize := len(dirList) / parts
	remainder := len(dirList) % parts

	start := 0

	for i := 0; i < parts; i++ {
		end := start + partSize
		if remainder > 0 {
			end++
			remainder--
		}

		result[i] = dirList[start:end]
		start = end
	}

	return result
}

// 重构：将文件路径转换为测试用例
func convertCaseProcess(paths []string) ([]*TestCase, error) {
	var cTestCases []*TestCase

	for _, path := range paths {
		testCasePath := TestCasePath(path)
		tc, err := testCasePath.ToTestCase()
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert TestCasePath to TestCase")
		}
		cTestCases = append(cTestCases, tc)
	}
	return cTestCases, nil
}

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

	/* 	if len(testCases) != 0 {
	   		testCasesList = append(testCasesList, testCases)
	   	}

	   	var count = 0

	   	var convertCaseProcess = func(paths []string) ([]*TestCase, error) {
	   		var cTestCases []*TestCase
	   		for _, path := range paths {
	   			count++
	   			testCasePath := TestCasePath(path)
	   			tc, err := testCasePath.ToTestCase()
	   			if err != nil {
	   				return nil, errors.Wrap(err, "failed to convert TestCasePath to TestCase")
	   			}
	   			cTestCases = append(cTestCases, tc)
	   		}
	   		return cTestCases, nil
	   	}

	   	splitDirList, err := splitDirArray(dirList, int(parallelism))
	   	if err != nil {
	   		return nil, errors.Wrap(err, "failed to split Dir parts")
	   	}

	   	for _, dl := range splitDirList {
	   		var testCasesT []*TestCase

	   		for _, dirName := range dl {
	   			paths := dirMap[dirName]
	   			tc, err := convertCaseProcess(paths)
	   			if err != nil {
	   				return nil, err
	   			}
	   			testCasesT = append(testCasesT, tc...)
	   		}

	   		testCasesList = append(testCasesList, testCasesT)
	   	}

	   	log.Info().Int("count", count).Msg("load testcases successfully") */
	return dirList, dirMap, nil

}

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

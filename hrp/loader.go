package hrp

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

func LoadTestCases(iTestCases ...ITestCase) ([][]*TestCase, error) {
	var testCasesList = make([][]*TestCase, 0)

	var dirMap = make(map[string][]string, 0)
	var dirList = make([]string, 0)

	for _, iTestCase := range iTestCases {
		/* if _, ok := iTestCase.(*TestCase); ok {
			testcase, err := iTestCase.ToTestCase()
			if err != nil {
				log.Error().Err(err).Msg("failed to convert ITestCase interface to TestCase struct")
				return nil, err
			}
			testCases = append(testCases, testcase)
			continue
		} */

		// iTestCase should be a TestCasePath, file path or folder path
		tcPath, ok := iTestCase.(*TestCasePath)
		if !ok {
			return nil, errors.New("invalid iTestCase type")
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
			return nil, errors.Wrap(err, "read dir failed")
		}

		for _, path := range paths {
			dirName := filepath.Dir(path)
			if _, ok := dirMap[dirName]; !ok {
				dirList = append(dirList, dirName)
			}
			dirMap[dirName] = append(dirMap[dirName], path)
		}

	}

	dirCount := len(dirList)
	ave := dirCount / 2
	var count = 0
	if ave == 0 {
		// 如果目录数量为1或0，则直接处理整个目录
		var testCases []*TestCase
		for _, dirName := range dirList {
			paths := dirMap[dirName]
			for _, path := range paths {
				count++
				testCasePath := TestCasePath(path)
				tc, err := testCasePath.ToTestCase()
				if err != nil {
					return nil, errors.Wrap(err, "failed to convert TestCasePath to TestCase")
				}
				testCases = append(testCases, tc)
			}
		}
		testCasesList = append(testCasesList, testCases)

	} else {
		for i := 0; i < dirCount; i += ave {
			var testCases []*TestCase

			end := i + ave
			if end > dirCount {
				end = dirCount
			}

			dirListTmp := dirList[i:end]
			for _, dirName := range dirListTmp {
				paths := dirMap[dirName]
				for _, path := range paths {
					count++
					testCasePath := TestCasePath(path)
					tc, err := testCasePath.ToTestCase()
					if err != nil {
						return nil,
							errors.Wrap(err, "failed to convert TestCasePath to TestCase")
					}
					testCases = append(testCases, tc)
				}
			}
			testCasesList = append(testCasesList, testCases)
		}
	}

	log.Info().Int("count", count).Msg("load testcases successfully")
	return testCasesList, nil
}

package mvc

import (
	"fmt"
	"html/template"
	"sync"
	"time"

	"github.com/zsy619/yyhertz/framework/util"
)

type Statistics struct {
	RequestURL        string
	RequestController string
	RequestNum        int64
	MinTime           time.Duration
	MaxTime           time.Duration
	TotalTime         time.Duration
}

type URLMap struct {
	lock        sync.RWMutex
	LengthLimit int // limit the urlmap's length if it's equal to 0 there's no limit
	urlmap      map[string]map[string]*Statistics
}

func (m *URLMap) AddStatistics(requestMethod, requestURL, requestController string, requesttime time.Duration) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if method, ok := m.urlmap[requestURL]; ok {
		if s, ok := method[requestMethod]; ok {
			s.RequestNum++
			if s.MaxTime < requesttime {
				s.MaxTime = requesttime
			}
			if s.MinTime > requesttime {
				s.MinTime = requesttime
			}
			s.TotalTime += requesttime
		} else {
			nb := &Statistics{
				RequestURL:        requestURL,
				RequestController: requestController,
				RequestNum:        1,
				MinTime:           requesttime,
				MaxTime:           requesttime,
				TotalTime:         requesttime,
			}
			m.urlmap[requestURL][requestMethod] = nb
		}
	} else {
		if m.LengthLimit > 0 && m.LengthLimit <= len(m.urlmap) {
			return
		}
		methodmap := make(map[string]*Statistics)
		nb := &Statistics{
			RequestURL:        requestURL,
			RequestController: requestController,
			RequestNum:        1,
			MinTime:           requesttime,
			MaxTime:           requesttime,
			TotalTime:         requesttime,
		}
		methodmap[requestMethod] = nb
		m.urlmap[requestURL] = methodmap
	}
}
func (m *URLMap) GetMap() map[string]interface{} {
	m.lock.RLock()
	defer m.lock.RUnlock()

	fields := []string{"requestUrl", "method", "times", "used", "max used", "min used", "avg used"}

	var resultLists [][]string
	content := make(map[string]interface{})
	content["Fields"] = fields

	for k, v := range m.urlmap {
		for kk, vv := range v {
			result := []string{
				fmt.Sprintf("% -50s", template.HTMLEscapeString(k)),
				fmt.Sprintf("% -10s", kk),
				fmt.Sprintf("% -16d", vv.RequestNum),
				fmt.Sprintf("%d", vv.TotalTime),
				fmt.Sprintf("% -16s", util.ToShortTimeFormat(vv.TotalTime)),
				fmt.Sprintf("%d", vv.MaxTime),
				fmt.Sprintf("% -16s", util.ToShortTimeFormat(vv.MaxTime)),
				fmt.Sprintf("%d", vv.MinTime),
				fmt.Sprintf("% -16s", util.ToShortTimeFormat(vv.MinTime)),
				fmt.Sprintf("%d", time.Duration(int64(vv.TotalTime)/vv.RequestNum)),
				fmt.Sprintf("% -16s", util.ToShortTimeFormat(time.Duration(int64(vv.TotalTime)/vv.RequestNum))),
			}
			resultLists = append(resultLists, result)
		}
	}
	content["Data"] = resultLists
	return content
}
func (m *URLMap) GetMapData() []map[string]interface{} {
	m.lock.RLock()
	defer m.lock.RUnlock()

	var resultLists []map[string]interface{}

	for k, v := range m.urlmap {
		for kk, vv := range v {
			result := map[string]interface{}{
				"request_url": k,
				"method":      kk,
				"times":       vv.RequestNum,
				"total_time":  util.ToShortTimeFormat(vv.TotalTime),
				"max_time":    util.ToShortTimeFormat(vv.MaxTime),
				"min_time":    util.ToShortTimeFormat(vv.MinTime),
				"avg_time":    util.ToShortTimeFormat(time.Duration(int64(vv.TotalTime) / vv.RequestNum)),
			}
			resultLists = append(resultLists, result)
		}
	}
	return resultLists
}

var StatisticsMap *URLMap

func init() {
	StatisticsMap = &URLMap{
		urlmap: make(map[string]map[string]*Statistics),
	}
}

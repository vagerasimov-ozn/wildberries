package communication // import "git.webdesk.ru/wd/kit/modules/communication"

//import "gopkg.in/webnice/debug.v1"
//import "gopkg.in/webnice/log.v2"
import (
	"bytes"
	"fmt"
	"strings"

	"gopkg.in/webnice/transport.v2"
	"gopkg.in/webnice/transport.v2/content"
	"gopkg.in/webnice/transport.v2/request"
	"gopkg.in/webnice/transport.v2/response"
	"gopkg.in/webnice/web.v1/header"
	"gopkg.in/webnice/web.v1/status"
)

func newTransport() transport.Interface {
	return transport.New().
		MaximumIdleConnections(defaultMaximumIdleConnections).               // Максимальное общее число бездействующих keepalive соединений
		MaximumIdleConnectionsPerHost(defaultMaximumIdleConnectionsPerHost). // Максимальное число бездействующих keepalive соединений для каждого хоста
		DialContextTimeout(defaultDialContextTimeout).                       // Таймаут установки соединения с хостом
		IdleConnectionTimeout(defaultIdleConnectionTimeout).                 // Таймаут keepalive соединения до обрыва связи
		TotalTimeout(defaultTotalTimeout).                                   // Общий таймаут на весь процесс связи, включает соединение, отправку данных, получение ответа
		RequestPoolSize(defaultRequestPoolSize)                              // Размер пула воркеров готовых для выполнения запросов к хостам
}

// Transport Готовый к использованию интерфейс коммуникации с сервером
func Transport() transport.Interface {
	if singleton == nil {
		singleton = newTransport()
	}
	return singleton
}

// RequestResponse Выполнение запроса, ожидание и получение результата
func RequestResponse(req request.Interface) (ret response.Interface, err error) {
	// DEBUG
	//req.DebugFunc(func(d []byte) { log.Debug(string(d)) })
	// DEBUG
	// Выполнение запроса
	singleton.Do(req)
	// Ожидание ответа
	if err = req.Done().Error(); err != nil {
		err = fmt.Errorf("execute request error: %s", err)
		return
	}
	// Анализ результата
	ret = req.Response()
	switch ret.StatusCode() {
	case status.Unauthorized:
		err = Errors().Unauthorized()
		return
	case status.Forbidden:
		err = Errors().Forbidden()
		return
	case status.NotFound:
		err = Errors().NotFound()
		return
	}
	if ret.StatusCode() < 200 || ret.StatusCode() > 299 {
		err = fmt.Errorf("request %s %q error, HTTP code %d (%s)", ret.Response().Request.Method, ret.Response().Request.URL.String(), ret.StatusCode(), ret.Status())
		return
	}

	return
}

// RequestResponseStatusCode Выполнение запроса, ожидание и получение результата в виде HTTP статуса
func RequestResponseStatusCode(req request.Interface) (ret int, err error) {
	var rsp response.Interface

	if rsp, err = RequestResponse(req); err != nil {
		return
	}
	ret = rsp.StatusCode()
	// DEBUG
	//req.Response().Content().BackToBegin()
	//log.Debug(req.Response().Content().String())
	// DEBUG

	return
}

// RequestResponsePlainText Выполнение запроса, ожидание и получение результата в виде текста
func RequestResponsePlainText(req request.Interface) (ret *bytes.Buffer, err error) {
	var (
		rsp response.Interface
		cnt content.Interface
	)

	if rsp, err = RequestResponse(req); err != nil {
		return
	}
	ret, cnt = &bytes.Buffer{}, rsp.Content()
	if strings.EqualFold(rsp.Header().Get(header.ContentEncoding), EncodingGzip) {
		cnt = cnt.UnGzip()
	}
	if strings.EqualFold(rsp.Header().Get(header.ContentEncoding), EncodingDeflate) {
		cnt = cnt.UnFlate()
	}
	_, err = cnt.WriteTo(ret)
	// DEBUG
	//req.Response().Content().BackToBegin()
	//log.Debug(req.Response().Content().String())
	// DEBUG

	return
}

// RequestResponseJSON Выполнение запроса, ожидание и получение результата в виде JSON
func RequestResponseJSON(req request.Interface, data interface{}) (err error) {
	var (
		rsp response.Interface
		cnt content.Interface
	)

	if rsp, err = RequestResponse(req); err != nil {
		return
	}
	cnt = rsp.Content()
	if strings.EqualFold(rsp.Header().Get(header.ContentEncoding), EncodingGzip) {
		cnt = cnt.UnGzip()
	}
	if strings.EqualFold(rsp.Header().Get(header.ContentEncoding), EncodingDeflate) {
		cnt = cnt.UnFlate()
	}
	err = cnt.UnmarshalJSON(data)
	// DEBUG
	//req.Response().Content().BackToBegin()
	//log.Debug(req.Response().Content().String())
	// DEBUG

	return
}

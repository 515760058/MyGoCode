import (
	"github.com/gorilla/websocket"
	"io"
	"sync"
	"time"
)

/**
 * 问题： 组件出现崩溃：panic: concurrent write to websocket connection
 * 原因： 当前使用websocket库 不支持并发操作. 在并发调用WriteMessage函数的时候导致崩溃
 * 参见：https://github.com/gorilla/websocket/issues/380

原文档：
	Connections support one concurrent reader and one concurrent writer.
	Concurrency Connections support one concurrent reader and one concurrent writer.
	Applications are responsible for ensuring that no more than one goroutine calls the write methods (NextWriter, SetWriteDeadline, WriteMessage, WriteJSON, EnableWriteCompression, SetCompressionLevel) concurrently
and that no more than one goroutine calls the read methods (NextReader, SetReadDeadline, ReadMessage, ReadJSON, SetPongHandler, SetPingHandler) concurrently.
	The Close and WriteControl methods can be called concurrently with all other methods.

 *
 * 修改：对websocket库再封装一层，加锁从而实现并发安全。
 */
type WebsocketConnection struct {
	writeMutex sync.Mutex // 限制1个写并发
	readMutex  sync.Mutex // 限制1个读并发
	wsConn     *websocket.Conn
}

// 构造函数
func NewWebsocketConnection(wsConn *websocket.Conn) *WebsocketConnection {
	if wsConn == nil {
		return nil
	}
	ws := &WebsocketConnection{}
	ws.wsConn = wsConn
	return ws
}

// 析构函数
func (wc *WebsocketConnection) Close()  {
	wc.wsConn.Close()
}


// 写并发
func (wc *WebsocketConnection) WriteControl(messageType int, data []byte, deadline time.Time) error {
	wc.writeMutex.Lock()
	defer wc.writeMutex.Unlock()
	return wc.wsConn.WriteControl(messageType, data, deadline)
}

func (wc *WebsocketConnection) WriteMessage(messageType int, data []byte) error {
	wc.writeMutex.Lock()
	defer wc.writeMutex.Unlock()
	return wc.wsConn.WriteMessage(messageType, data)
}

func (wc *WebsocketConnection) WriteJSON(v interface{}) error {
	wc.writeMutex.Lock()
	defer wc.writeMutex.Unlock()
	return wc.wsConn.WriteJSON(v)
}

func (wc *WebsocketConnection) EnableWriteCompression(enable bool) {
	wc.writeMutex.Lock()
	defer wc.writeMutex.Unlock()
	wc.wsConn.EnableWriteCompression(enable)
}

func (wc *WebsocketConnection) NextWriter(messageType int) (io.WriteCloser, error)  {
	wc.writeMutex.Lock()
	defer wc.writeMutex.Unlock()
	return wc.wsConn.NextWriter(messageType)
}

func (wc *WebsocketConnection) SetWriteDeadline(t time.Time) error  {
	wc.writeMutex.Lock()
	defer wc.writeMutex.Unlock()
	return wc.wsConn.SetWriteDeadline(t)
}



// 读并发
func (wc *WebsocketConnection) NextReader() (messageType int, r io.Reader, err error) {
	wc.readMutex.Lock()
	defer wc.readMutex.Unlock()
	return wc.wsConn.NextReader()
}

func (wc *WebsocketConnection) SetReadDeadline(t time.Time) error{
	wc.readMutex.Lock()
	defer wc.readMutex.Unlock()
	return wc.wsConn.SetReadDeadline(t)
}

func (wc *WebsocketConnection) ReadMessage() (messageType int, p []byte, err error) {
	wc.readMutex.Lock()
	defer wc.readMutex.Unlock()
	return wc.wsConn.ReadMessage()
}

func (wc *WebsocketConnection) ReadJSON(v interface{}) error{
	wc.readMutex.Lock()
	defer wc.readMutex.Unlock()
	return wc.wsConn.ReadJSON(v)
}

func (wc *WebsocketConnection) SetPongHandler(h func(appData string) error) {
	wc.readMutex.Lock()
	defer wc.readMutex.Unlock()
	wc.wsConn.SetPongHandler(h)
}

func (wc *WebsocketConnection) SetPingHandler(h func(appData string) error) {
	wc.readMutex.Lock()
	defer wc.readMutex.Unlock()
	wc.wsConn.SetPingHandler(h)
}




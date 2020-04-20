package tcp

/*
	TCP传输规定
	格式： op<keyLen><sp><vauleLen><sp><key><value>
	op表示操作，S|G|D 分别表示Set操作，Get操作，Del操作
	新增操作  T3 5 2 <key><value><expiretime>带过期时间的Set
	<sp>表示空格
	<keyLen>表示传输的key的字长
	例如： S3 5 keyvalue 表示Set key vaule
          G3 key 表示获取key的value
		  D3 key 表示删除key的内容
*/
import (
	"Godis/cache"
	"Godis/cluster"
	"Godis/config"
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
)

type Server struct {
	cache.Cache
	cluster.Node
}

type result struct {
	val []byte
	err error
}

func NewServer(c cache.Cache, node cluster.Node) *Server {
	return &Server{c, node}
}

func (s *Server) Listen(port string, ctx context.Context) {
	addr := fmt.Sprintf("%s:%s", strings.Split(s.Addr(), ":")[0], port)
	listener, err := net.Listen("tcp", addr)
	//listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log.Printf("监听端口%v失败\nerror:%v\n", port, err)
		return
	}
	//log.Printf("tcp监听地址：%s\n",addr)
	go func() {
		select {
		case <-ctx.Done():
			listener.Close()
		}
	}()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf(err.Error())
			continue
		}
		//log.Printf("客户端连接成功 %v\n",conn.RemoteAddr())
		go s.process(conn)
	}
}

//异步处理，请求密集时能有效提高性能
func (s *Server) processWithAsync(conn net.Conn) {
	r := bufio.NewReader(conn)
	resultCh := make(chan chan *result, 5000)
	defer close(resultCh)
	go reply(conn, resultCh)
	for {
		op, err := r.ReadByte()
		if err != nil {
			if err != io.EOF {
				//log.Printf("连接已关闭错误：%v\n",err.Error())
			}
			return
		}
		switch op {
		case 'S':
			s.setWithAsync(resultCh, r)
		case 'G':
			s.getWithAsync(resultCh, r)
		case 'D':
			s.delWithAsync(resultCh, r)
		case 'T':
			s.setWithTimeWithAsync(resultCh, r)
		default:
			log.Printf("错误操作%v\n", op)
			return
		}
	}
}

func reply(conn net.Conn, resultCh chan chan *result) {
	defer conn.Close()
	for {
		c, open := <-resultCh
		if !open {
			return
		}
		r := <-c
		err := sendResponse(r.val, conn, r.err)
		if err != nil {
			//log.Printf("连接已关闭错误：%v\n",err.Error())
			return
		}
	}
}

func (s *Server) process(conn net.Conn) {
	defer conn.Close()
	//defer log.Printf("客户端断开连接 %v\n",conn.RemoteAddr())
	reader := bufio.NewReader(conn)
	for {
		op, err := reader.ReadByte()
		if err != nil {
			if err != io.EOF {
				//log.Printf("连接已关闭错误：%v\n",err.Error())
			}
			return
		}
		switch op {
		case 'S':
			err = s.set(conn, reader)
		case 'G':
			err = s.get(conn, reader)
		case 'D':
			err = s.del(conn, reader)
		case 'T':
			err = s.setWithTime(conn, reader)
		default:
			log.Printf("错误操作%v\n", string(op))
			return
		}
		if err != nil {
			log.Printf(err.Error())
			return
		}
	}

}

func readLen(r *bufio.Reader) (int, error) {
	tempLen, err := r.ReadString(' ')
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.TrimSpace(tempLen))
}

func (s *Server) readKey(r *bufio.Reader) (string, error) {
	keyLen, err := readLen(r)
	if err != nil {
		return "", err
	}
	key := make([]byte, keyLen)
	n, err := io.ReadFull(r, key)
	if err != nil || n != keyLen {
		return "", err
	}
	k := string(key)
	addr, ok := s.ShouldProcess(k)
	if !ok {
		return "", errors.New("redirect:" + addr)
	}
	return k, nil
}

func (s *Server) readKeyAndValue(r *bufio.Reader) (string, []byte, error) {
	keyLen, err := readLen(r)
	if err != nil {
		return "", nil, err
	}
	valueLen, err := readLen(r)
	if err != nil {
		return "", nil, err
	}
	key := make([]byte, keyLen)
	value := make([]byte, valueLen)
	n, err := io.ReadFull(r, key)
	if err != nil || n != keyLen {
		return "", nil, err
	}
	k := string(key)
	n, err = io.ReadFull(r, value)
	if err != nil || n != valueLen {
		return "", nil, err
	}
	//先读完bufio中的内容，再验证是否应该是当前节点处理，防止后续服务端读到错误操作内容
	addr, ok := s.ShouldProcess(k)
	if !ok {
		return "", nil, errors.New("redirect:" + addr)
	}
	return k, value, nil
}

//响应查询结果
//'-'开头表示出现异常,则写入<-><errLen><sp><errMessage>
//正常应该直接写入<valueLen><sp><value>
func sendResponse(value []byte, conn net.Conn, err error) error {
	if err != nil {
		_, err := conn.Write([]byte(fmt.Sprintf("-%d %s", len(err.Error()), err.Error())))
		return err
	}
	conn.Write([]byte(fmt.Sprintf("%d ", len(value))))
	_, err = conn.Write(value)
	return err
}

func (s *Server) get(conn net.Conn, r *bufio.Reader) error {
	key, err := s.readKey(r)
	if err != nil {
		return err
	}
	values, err := s.Get(key)
	return sendResponse(values, conn, err)
}

func (s *Server) set(conn net.Conn, r *bufio.Reader) error {
	key, val, err := s.readKeyAndValue(r)
	if err != nil && strings.Contains(err.Error(), "redirect") {
		return sendResponse(nil, conn, err)
	}
	if err != nil {
		return err
	}
	return sendResponse(nil, conn, s.Set(key, cache.Value{val, time.Now(), time.Second * time.Duration(config.Config.ExpireStrategy.DefaultExpireTime)}))
}

func (s *Server) setWithTime(conn net.Conn, r *bufio.Reader) error {
	key, val, expireTime, err := s.readKeyAndValueAndTime(r)
	if err != nil {
		return err
	}
	return sendResponse(nil, conn, s.Set(key, cache.Value{val, time.Now(), time.Second * expireTime}))
}

func (s *Server) del(conn net.Conn, r *bufio.Reader) error {
	key, err := s.readKey(r)
	if err != nil {
		return err
	}
	return sendResponse(nil, conn, s.Del(key))
}

func (s *Server) getWithAsync(ch chan chan *result, r *bufio.Reader) {
	c := make(chan *result)
	ch <- c
	key, err := s.readKey(r)
	if err != nil {
		c <- &result{nil, err}
		return
	}
	go func() {
		val, err := s.Get(key)
		c <- &result{val, err}
	}()
}

func (s *Server) setWithAsync(ch chan chan *result, r *bufio.Reader) {
	c := make(chan *result)
	ch <- c
	key, val, err := s.readKeyAndValue(r)
	if err != nil {
		c <- &result{nil, err}
		return
	}
	go func() {
		c <- &result{nil, s.Set(key, cache.Value{val, time.Now(), 0})}
	}()
}

func (s *Server) setWithTimeWithAsync(ch chan chan *result, r *bufio.Reader) {
	c := make(chan *result)
	ch <- c
	key, val, expireTime, err := s.readKeyAndValueAndTime(r)
	if err != nil {
		c <- &result{nil, err}
		return
	}
	go func() {
		c <- &result{nil, s.Set(key, cache.Value{val, time.Now(), time.Second * expireTime})}
	}()
}

func (s *Server) delWithAsync(ch chan chan *result, r *bufio.Reader) {
	c := make(chan *result)
	ch <- c
	key, err := s.readKey(r)
	if err != nil {
		c <- &result{nil, err}
		return
	}
	go func() {
		c <- &result{nil, s.Del(key)}
	}()
}

func (s *Server) readKeyAndValueAndTime(r *bufio.Reader) (string, []byte, time.Duration, error) {
	keyLen, err := readLen(r)
	if err != nil {
		return "", nil, -1, err
	}
	valueLen, err := readLen(r)
	if err != nil {
		return "", nil, -1, err
	}
	timeLen, err := readLen(r)
	if err != nil {
		return "", nil, -1, err
	}
	key := make([]byte, keyLen)
	value := make([]byte, valueLen)
	t := make([]byte, timeLen)
	n, err := io.ReadFull(r, key)
	if err != nil || n != keyLen {
		return "", nil, -1, err
	}
	n, err = io.ReadFull(r, value)
	if err != nil || n != valueLen {
		return "", nil, -1, err
	}
	n, err = io.ReadFull(r, t)
	if err != nil || n != timeLen { //修复判断长度变量错误
		return "", nil, -1, err
	}
	//先读完bufio中的内容，再验证是否应该是当前节点处理，防止后续服务端读到错误操作内容
	k := string(key)
	addr, ok := s.ShouldProcess(k)
	if !ok {
		return "", nil, 0, errors.New("redict:" + addr)
	}
	expireTime, err := strconv.Atoi(string(t))
	if err != nil || expireTime < 0 {
		return "", nil, -1, err
	}
	return k, value, time.Duration(expireTime), nil
}

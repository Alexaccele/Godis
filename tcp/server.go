package tcp
/*
	TCP传输规定
	格式： op<keyLen><sp><vauleLen><sp><key><value>
	op表示操作，S|G|D 分别表示Set操作，Get操作，Del操作
	<sp>表示空格
	<keyLen>表示传输的key的字长
	例如： S3 5 keyvalue 表示Set key vaule
          G3 key 表示获取key的value
		  D3 key 表示删除key的内容
 */
import (
	"Godis/cache"
	"bufio"
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
}

type result struct{
	val []byte
	err error
}

func NewServer(c cache.Cache) *Server {
	return &Server{c}
}

func (s *Server) Listen(port string)  {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	//listener, err := net.ListenTCP("tcp", addr)
	if err!=nil{
		log.Printf("监听端口%v失败\nerror:%v\n",port,err)
		return
	}
	for{
		conn, err := listener.Accept()
		if err != nil{
			log.Printf(err.Error())
			continue
		}
		log.Printf("客户端连接成功 %v\n",conn.RemoteAddr())
		go s.process(conn)
	}
}
//异步处理，请求密集时能有效提高性能
func (s *Server) processWithAsync(conn net.Conn){
	r := bufio.NewReader(conn)
	resultCh := make(chan chan *result,5000)
	defer close(resultCh)
	go reply(conn,resultCh)
	for{
		op, err := r.ReadByte()
		if err!=nil{
			if err!=io.EOF{
				log.Printf("连接已关闭错误：%v\n",err.Error())
			}
			return
		}
		switch op {
		case 'S':
			s.setWithAsync(resultCh,r)
		case 'G':
			s.getWithAsync(resultCh,r)
		case 'D':
			s.delWithAsync(resultCh,r)
		default:
			log.Printf("错误操作%v\n",op)
			return
		}
	}
}

func reply(conn net.Conn,resultCh chan chan *result){
	defer conn.Close()
	for{
		c,open := <- resultCh
		if !open{
			return
		}
		r := <-c
		err := sendResponse(r.val,conn,r.err)
		if err != nil{
			log.Printf("连接已关闭错误：%v\n",err.Error())
			return
		}
	}
}

func (s *Server) process(conn net.Conn)  {
	defer conn.Close()
	defer log.Printf("客户端断开连接 %v\n",conn.RemoteAddr())
	reader := bufio.NewReader(conn)
	for{
		op, err := reader.ReadByte()
		if err!=nil{
			if err!=io.EOF{
				log.Printf("连接已关闭错误：%v\n",err.Error())
			}
			return
		}
		switch op {
		case 'S':
			err = s.set(conn,reader)
		case 'G':
			err = s.get(conn,reader)
		case 'D':
			err = s.del(conn,reader)
		default:
			log.Printf("错误操作%v\n",op)
			return
		}
		if err!=nil{
			log.Printf(err.Error())
			return
		}
	}

}

func readLen(r *bufio.Reader) (int,error) {
	tempLen, err := r.ReadString(' ')
	if err!=nil{
		return 0,err
	}
	return strconv.Atoi(strings.TrimSpace(tempLen))
}

func (s *Server) readKey(r *bufio.Reader) (string,error) {
	keyLen, err := readLen(r)
	if err!=nil{
		return "",err
	}
	key:=make([]byte,keyLen)
	n, err := io.ReadFull(r,key)
	if err!=nil || n!=keyLen{
		return "",err
	}
	return string(key),nil
}

func (s *Server) readKeyAndValue(r *bufio.Reader) (string,[]byte,error) {
	keyLen, err := readLen(r)
	if err!=nil{
		return "",nil,err
	}
	valueLen, err := readLen(r)
	if err!=nil{
		return "",nil,err
	}
	key := make([]byte,keyLen)
	value := make([]byte,valueLen)
	n, err := io.ReadFull(r, key)
	if err!=nil||n!=keyLen{
		return "",nil,err
	}
	n, err = io.ReadFull(r, value)
	if err!=nil||n!=valueLen{
		return "",nil,err
	}
	return string(key),value,nil
}
//响应查询结果
//'-'开头表示出现异常,则写入<-><errLen><sp><errMessage>
//正常应该直接写入<valueLen><sp><value>
func sendResponse(value []byte,conn net.Conn,err error) error {
	if err!=nil{
		conn.Write([]byte(fmt.Sprintf("-%d ")))
		_, err = conn.Write([]byte(err.Error()))
		return err
	}
	conn.Write([]byte(fmt.Sprintf("%d ",len(value))))
	_, err = conn.Write(value)
	return err
}

func (s *Server) get(conn net.Conn,r *bufio.Reader) error {
	key, err := s.readKey(r)
	if err!=nil{
		return err
	}
	values, err := s.Get(key)
	return sendResponse(values,conn,err)
}

func (s *Server) set(conn net.Conn,r *bufio.Reader) error {
	key, val,err := s.readKeyAndValue(r)
	if err!=nil {
		return err
	}
	return sendResponse(nil,conn,s.Set(key,cache.Value{val,time.Now(),0}))
}

func (s *Server) del(conn net.Conn,r *bufio.Reader) error {
	key, err := s.readKey(r)
	if err!=nil{
		return err
	}
	return sendResponse(nil,conn,s.Del(key))
}

func (s *Server) getWithAsync(ch chan chan *result,r *bufio.Reader){
	c := make(chan *result)
	ch <- c
	key,err := s.readKey(r)
	if err != nil{
		c <- &result{nil,err}
		return
	}
	go func(){
		val,err := s.Get(key)
		c <- &result{val,err};
	}()
}

func (s *Server) setWithAsync(ch chan chan *result,r *bufio.Reader){
	c := make(chan *result)
	ch <- c
	key, val,err := s.readKeyAndValue(r)
	if err != nil{
		c <- &result{nil,err}
		return
	}
	go func(){
		c <- &result{nil,s.Set(key,cache.Value{val,time.Now(),0})};
	}()
}

func (s *Server) delWithAsync(ch chan chan *result,r *bufio.Reader){
	c := make(chan *result)
	ch <- c
	key,err := s.readKey(r)
	if err != nil{
		c <- &result{nil,err}
		return
	}
	go func(){
		c <- &result{nil,s.Del(key)};
	}()
}
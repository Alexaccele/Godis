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
	"net"
	"strconv"
	"strings"
)

type Server struct {
	cache.Cache
}

func NewServer(c cache.Cache) *Server {
	return &Server{c}
}

func (s *Server) Listen(port string)  {
	addr, _ := net.ResolveTCPAddr("tcp", fmt.Sprintf(":%s", port))
	listener, err := net.ListenTCP("tcp", addr)
	if err!=nil{
		fmt.Printf("监听端口%v失败",port)
		return
	}
	for{
		conn, err := listener.Accept()
		if err != nil{
			fmt.Printf(err.Error())
			continue
		}
		go s.process(conn)
	}
}

func (s *Server) process(conn net.Conn)  {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	for{
		op, err := reader.ReadByte()
		if err!=nil{
			if err!=io.EOF{
				fmt.Printf("连接关闭导致错误：%v\n",err)
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
			fmt.Printf("错误操作%v\n",op)
			return
		}
		if err!=nil{
			fmt.Printf(err.Error())
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
	writer := bufio.NewWriter(conn)
	if err!=nil{
		writer.Write([]byte(fmt.Sprintf("-%d ")))
		writer.Write([]byte(err.Error()))
		return writer.Flush()
	}
	writer.Write([]byte(fmt.Sprintf("%d ",len(value))))
	writer.Write(value)
	return writer.Flush()
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
	key, value,err := s.readKeyAndValue(r)
	if err!=nil {
		return err
	}
	return sendResponse(nil,conn,s.Set(key,value))
}

func (s *Server) del(conn net.Conn,r *bufio.Reader) error {
	key, err := s.readKey(r)
	if err!=nil{
		return err
	}
	return sendResponse(nil,conn,s.Del(key))
}
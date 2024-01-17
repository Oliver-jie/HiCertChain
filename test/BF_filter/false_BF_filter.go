package main
import(
	"fmt"
	"math/rand"
	"strconv"
	"io"
	"crypto/md5"
	"bytes"
	"encoding/binary"
	"time"
)
// 双重布谷鸟过滤器代码
func main () {

	// table_length := 1048576

	var table [1048576]int


	//第n次的插入
	times := 0


	//开始装填证书信息
	
	for i:=1;i<=62259;i++ { 
		times++
		fmt.Printf("开始第%d次插入",times)
		fmt.Printf("---\n")

		for j:=1;j<=16;j++{
			table[compu_chart(i,j)]=1
		}
	}
	time1:=0
	find_times:=0
	temp_time:=0
	find_right:=true
	// 假阳性测试 100000-1100000 200000-1200000 300000-1300000 400000-1400000 500000-1500000
	start := time.Now()
	for m:=500000;m<=1500000;m++ {
		temp_time=0
		find_right=true
		time1++
		fmt.Printf("开始第%d次查询\n",time1)
        fmt.Printf("---\n")
		for n:=1;n<=16;n++{
			if table[compu_chart(m,n)]!=1{
				find_right=false
				break
			}
			temp_time++
		}
		if temp_time==16{
			find_times++
		}
	}
	
	elapsed := time.Since(start).Nanoseconds()
	find_right=true
	if find_right==true{
		fmt.Printf("假阳性查询的次数为%d\n",find_times)
		fmt.Printf("代码运行的总时间为%d\n",elapsed)
	}
	// fmt.Printf("插入失败的次数为%d\n",MAX_colli_time)

	
}

// 计算要放入的位置
func compu_chart(cert0String int,num int)int{
	certString :=MD5Hash2(strconv.Itoa(cert0String))
	cert1String := certString[num:(num+8)]
	return (Hex2Dec2(cert1String))%1048576
}


func MD5Hash2(data string) string {
    t := md5.New()
    io.WriteString(t, data)
    return fmt.Sprintf("%x", t.Sum(nil))
}

func Int64ToBytes2(num int64) []uint8 {
	var buffer bytes.Buffer
	err := binary.Write(&buffer, binary.BigEndian, num)
	if err != nil {
		fmt.Println("int64转[]uint8失败")
	}
	return buffer.Bytes()
   }

func Hex2Dec2(val string) int {
	n, err := strconv.ParseUint(val, 16, 32)
	if err != nil {
		fmt.Println(err)
	}
	return int(n)
}
func RandInt(len int)int  {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return r.Intn(len)
	
}

func RandString2(len int) string {
    r := rand.New(rand.NewSource(time.Now().UnixNano()))
    bytes := make([]byte, len)
    for i := 0; i < len; i++ {
        b := r.Intn(26) + 65
        bytes[i] = byte(b)
    }
    return string(bytes)
}

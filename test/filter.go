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

func main () {
	//开始设计两张表分别为表一和表二，设计表的大小规则
	// 定义数据
	MAX_colli:= 16
	MAX_colli_time := 0 
	total_colli := 0
	temp_colli := 0
	//table_len := 100
	table_bread := 100
	var temp_data string

	var table1 [100][100]string
	var table1_num [100]int
	var table2 [100][100]string
	var table2_num [100]int
	time := 0

	//开始装填证书信息
	for i:=0;i<=9000;i++ {
		time++
		fmt.Printf("%d",time)
		fmt.Printf("---")
		//随机生成一个证书编号，是数字加随机字符串
		cert1:=RandInt(10000000)
		cert1String := strconv.Itoa(cert1) + RandString2(6)
		fmt.Printf(cert1String)
		fmt.Printf("---")
		
		//计算随机哈希，转换成int类型
		cert2Hash :=MD5Hash2(cert1String)		
		cert2String := cert2Hash[0:8]
		// cert2 := Hex2Dec2(cert2String)


		// 计算放入表的位置，一共有两个，一个是chart1,一个是chart2
		chart1 := compu_chart(cert1String,1)
		// 放入第二章表的数据
		chart2 := compu_chart(cert1String,2)
		fmt.Printf("%d",chart1)
		fmt.Printf("---")
		fmt.Printf("%d",chart2)
		fmt.Printf("---")
		fmt.Printf("\n")
		

		next_table1 := true
		temp_colli = 0

		//循环放入两张表中
		for j:=0;j<MAX_colli/2;j++{
			//放入第一张表中
		    if (table1_num[chart1]<table_bread && next_table1 == true && temp_colli <= MAX_colli){
				table1[chart1][table1_num[chart1]] = cert1String
				if table1_num[chart1]< 100 {
					table1_num[chart1]++
				}
				next_table1 = true
				temp_colli = 0
				break
			}else if (table1_num[chart1] == table_bread && next_table1 == true && temp_colli <= MAX_colli){
				// 第一张表满了，放入第二章表中
				//r=rand.New(rand.NewSource(time.Now().UnixNano()))
				randplace := RandInt(100)
				// randplace := rand.Intn(100)
				temp_data = table1[chart1][randplace]
				if len(temp_data)==0 {
					fmt.Println("cishujvshikong")
					fmt.Println(randplace)
				}
				fmt.Printf("\n")
				fmt.Println(temp_data)
				table1[chart1][randplace] = cert1String
				chart2 = compu_chart(temp_data,2)
				cert1String = temp_data
				next_table1 = false
				temp_colli++
				total_colli++
			}else if (table2_num[chart2]<table_bread && next_table1 == false && temp_colli <= MAX_colli){
				// 放入第二章表中
				table2[chart2][table2_num[chart2]] = cert1String
				if table2_num[chart2]< 100 {
					table2_num[chart2]++
				}
				next_table1 = true
				temp_colli = 0
				break
			}else if (table2_num[chart2] ==table_bread && next_table1 == false && temp_colli <= MAX_colli) {
				// r =rand.New(rand.NewSource(time.Now().UnixNano()))
				randplace := RandInt(100)
				// randplace := rand.Intn(100)
				temp_data = table2[chart2][randplace]
				table2[chart2][randplace] = cert1String
				chart1 = compu_chart(temp_data,1)
				cert1String = temp_data
				next_table1 = true
				temp_colli++
				total_colli++
			}else{
				MAX_colli_time++
				break
			}
		}
		
	}
	// 输出表的数据
	mm:=0
	nn:=0
	for m := 0; m < 100; m++ {
		fmt.Printf("%d",table1_num[m])
		fmt.Printf("--")
		if table1_num[m]==100 {
			mm++
		}
	}
	for n := 0; n < 100; n++ {
		fmt.Printf("%d",table2_num[n])
		fmt.Printf("--")
		if table2_num[n]==100 {
			nn++
		}
	}
	fmt.Printf("\n")
	fmt.Println(mm)
	fmt.Println(nn)
	fmt.Printf("发生总共碰撞的此时为%d",total_colli)
	fmt.Printf("\n")
	fmt.Printf("总共花费的时间是：\n",elapsed)
}
// 计算要放入的位置
func compu_chart(cert1String string,num int)int{
	cert1,_ := strconv.Atoi(cert1String[0:len(cert1String)-6])
	cert2Hash :=MD5Hash2(cert1String)
	cert2String := cert2Hash[0:8]
	cert2 := Hex2Dec2(cert2String)

	if num == 1 {
		chart := cert1%99
		return chart
	}else{
		chart := cert1%99
		chart2 := (cert2^chart)%99
		return chart2
	}
	

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

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
	"strings"
)
// 双重布谷鸟过滤器代码
func main () {
	//开始设计两张表分别为表一和表二，设计表的大小规则
	// 定义数据
	//最大的碰撞次数
	MAX_colli:= 172  
	//最大的碰撞时间  
	MAX_colli_time := 0 
	//一次插入总共的碰撞次数
	total_colli := 0
	//临时的一次碰撞
	temp_colli := 0
	//表的长度
	//table_len := 100
	//表的宽度
	table_bread := 4
	//table_high := 2048
	//临时的数据变量
	var temp_data string
	//插入次数计算
	insert_time := 0
	//表1的结构
	var table1 [2048][16]string
	//表一的附属表
	var table1_num [2048]int
	//表2的结构
	var table2 [2048][16]string
	//表2附属表
	var table2_num [2048]int
	//第n次的插入
	time1 := 0

	//证书插入 95%-62259  
	//开始装填证书信息
	start := time.Now()
	for i:=0;i<=62259;i++ {
		//开始第一次插入
		time1++
		fmt.Printf("%d",time1)
		fmt.Printf("---")

		// 计算放入表的位置，一共有两个，一个是row1,一个是row2
		row1,_ := strconv.Atoi(compu_chart(strconv.Itoa(i),1))
		row2,_ := strconv.Atoi(compu_chart(strconv.Itoa(i),2))
		// row2,_ := strconv.Atoi(compu_chart(cert1String,2))
		fpString := compu_chart(strconv.Itoa(i),3)
		fmt.Printf("%d",row1)
		fmt.Printf("---")
		fmt.Printf("\n")
		
		next_table1 := true
		temp_colli = 0

		//循环放入两张表中，次数不能超过最大的碰撞次数
		for j:=0;j<MAX_colli/2;j++{
			//放入第一张表中
			// 表一的该行没有满，接下来是放入表一，还没有达到最大的碰撞次数
		    if (table1_num[row1]<table_bread && next_table1 == true ){
				// 表1中 该行没有满
				if table1[row1][table1_num[row1]] =="" {
					table1[row1][table1_num[row1]] = fpString
					table1_num[row1]++
					next_table1 = true
					temp_colli = 0
					break
				// 解决因删除问题造成的空洞
				}else{
					// 找到空洞的位置
					for  table1[row1][insert_time]!=""{
						insert_time++
					}
					if table1[row1][insert_time]!="" {
						 fmt.Println("插入失败，请查询原因")
						for k := 0; k < 100; k++ {
							 fmt.Println("插入错误，请检查")
						}
						break
					}else{
						// 找到了空洞的位置
						table1[row1][insert_time] = fpString
						table1_num[row1]++
						next_table1 = true
						temp_colli = 0
						insert_time = 0
						break
					}
				}
			// 表1满了，先放入表一，再碰撞放入表2，还未达到最大碰撞次数
			}else if (table1_num[row1] == table_bread && next_table1 == true){
				randplace := RandInt(table_bread)
				temp_data = table1[row1][randplace]
				// 错误验证
				if len(temp_data)==0 {
					 fmt.Println("此处数据为空\n")
					 fmt.Println(randplace)
					 fmt.Println(temp_data)
				}
				//将数据插入表中
				table1[row1][randplace] = fpString
				//放置临时数据，更新插入位置，插入表，插入数据，碰撞次数
				row2 = fptochart(temp_data,row1)
				next_table1 = false
				fpString = temp_data
				temp_colli++
				total_colli++
			// 开始放入第二表中，首先第二章表未满，轮到第二张表；未达到最大碰撞次数
			}else if (table2_num[row2]<table_bread && next_table1 == false ){
				// 第二章表为空
				if table2[row2][table2_num[row2]] =="" {
					table2[row2][table2_num[row2]] = fpString
					table2_num[row2]++
					next_table1 = true
					temp_colli = 0
					break
				}else{
					for  table2[row2][insert_time] !=""{
						insert_time++						
					}
					if table2[row2][insert_time] !="" {
						fmt.Println("插入错误，请检查")
						for j := 0; j < 100; j++ {
							fmt.Println("插入错误，请检查")
						}
						break
					}else{
						table2[row2][insert_time] = fpString
						table2_num[row2]++
						next_table1 = true
						temp_colli = 0
						break
					}
				}
			// 开始放入第二表中，首先第二章表已满，轮到第二张表；未达到最大碰撞次数
			}else if (table2_num[row2] ==table_bread && next_table1 == false) {
				randplace := RandInt(table_bread)
				temp_data = table2[row2][randplace]
				table2[row2][randplace] = fpString
				// 准备插入临时数据
				row1 = fptochart(temp_data,row2)
				next_table1 = true
				fpString = temp_data
				temp_colli++
				total_colli++
			}else{
				MAX_colli_time++
				break
			}
		}
		
	}
	elapsed := time.Since(start).Nanoseconds()

	right_num:=0
    // 查询的速度
	for t:=500000;t<=562259;t++{
		string3:=strconv.Itoa(t)
		right_true:=false
		fmt.Printf("查询次数为%d\n",t)
		chart1,_ := strconv.Atoi(compu_chart(string3,1))
		// 放入第二章表的数据
		chart2,_ := strconv.Atoi(compu_chart(string3,2))
		string3 = compu_chart(string3,3)
		for k:=0;k<16;k++{
			if strtofp(table1[chart1][k])==strtofp(string3)||strtofp(table2[chart2][k])==strtofp(string3) {
				right_num++
				right_true = true
				break
			}
		}
		if right_true == false{
			fmt.Printf("%d在%d和%d查询不到%s\n",t,chart1,chart2,string3)
		}
	}

	fmt.Printf("假阳性的次数为%d\n",right_num)
	fmt.Printf("插入错误的次数为%d\n",MAX_colli_time)
	fmt.Printf("发生总共碰撞的此时为%d\n",total_colli)
	fmt.Printf("代码运行的总时间为%d\n",elapsed)

	
}
// 计算要放入的位置
func compu_chart(cert1String string,num int)string{
	cert2Hash :=MD5Hash2(cert1String)
	// hash截取
	cert3String := cert2Hash[2:10]
	// 转换成十进制
	chart:=Hex2Dec2(cert3String)%2048
	// 指纹信息
	cert2String := cert2Hash[0:8]
	fp := (Hex2Dec2(cert2String))%16384
	row2 := fp^chart
	if num == 1 {
		return strconv.Itoa(chart)
	}
	if num == 2{
		return strconv.Itoa(row2%2048)
	}
	return strconv.Itoa(fp)+"--"+strconv.Itoa(chart)+"--"+strconv.Itoa(row2)
}

func fptochart(fp string,chart int)int {
	certinfo:=strings.Split(fp,"--")
	info1,_:=strconv.Atoi(certinfo[1])
	info2,_:=strconv.Atoi(certinfo[2])
	if chart == info1 {
		return info2%2048
	}
	return info1%2048
}

func strtofp(fp string) string {
	certinfo:=strings.Split(fp,"--")
	return certinfo[0]
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

// func deleteOnedata(table1 [2500][4]string,table2 [2500][4]string,table1_num [2500]int,table2_num [2500]int)  {
// 	del_sign := 1
// 	for  del_sign ==1{
// 		del0 := RandInt(100)
// 		del1 := RandInt(2500)
// 		del2 := RandInt(4)
// 		//删除存在空洞
// 		if del0%2 ==0{
// 			if table1[del1][del2]!="" {
// 				table1[del1][del2] = ""
// 				del_sign =0
// 				table1_num[del1]--
// 			}	
// 		}else{
// 			if table2[del1][del2]!="" {
// 				table2[del1][del2] = ""
// 				del_sign=0
// 				table2_num[del1]--
// 			}
// 		}
// 	}		
// }
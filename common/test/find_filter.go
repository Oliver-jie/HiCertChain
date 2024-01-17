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
	//开始设计两张表分别为表一和表二，设计表的大小规则
	// 定义数据
	//最大的碰撞次数
	MAX_colli:= 16  
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
	//临时的数据变量
	var temp_data string
	//插入次数计算
	insert_time := 0
	//表1的结构
	var table1 [2500][4]string
	//表一的附属表
	var table1_num [2500]int
	//表2的结构
	var table2 [2500][4]string
	//表2附属表
	var table2_num [2500]int
	//第n次的插入
	time1 := 0

	//开始装填证书信息
	
	for i:=0;i<=19000;i++ {
		// 删除机制，随机删除一个数据
		// if i>=8000 {
		// 	deleteOnedata(table1,table2,table1_num,table2_num)
		// }
		//开始第一次插入
		time1++
		fmt.Printf("time1是%d",time1)
		fmt.Printf("---")
		//随机生成一个证书编号，是数字加随机字符串
		// cert1:=RandInt(1000000)
		//cert1String := strconv.Itoa(cert1) + RandString2(6)
		cert1String := strconv.Itoa(i*i)
		fmt.Printf("cert1String是%s",cert1String)
		fmt.Printf("---")
		
		//计算随机哈希，转换成int类型
		cert2Hash :=MD5Hash2(cert1String)		
		cert2String := cert2Hash[0:6]
		fmt.Printf(cert2String)
		// cert2 := Hex2Dec2(cert2String)


		// 计算放入表的位置，一共有两个，一个是chart1,一个是chart2
		chart1 := compu_chart(cert1String,1)
		// 放入第二章表的数据
		chart2 := compu_chart(cert1String,2)
		
		cert1String = strconv.Itoa(compu_chart(cert1String,3))
		fmt.Printf("chart1是%d",chart1)
		fmt.Printf("---")
		fmt.Printf("chart2是%d",chart2)
		fmt.Printf("---")
		fmt.Printf("\n")
		
		next_table1 := true
		temp_colli = 0

		//循环放入两张表中，次数不能超过最大的碰撞次数
		for j:=0;j<MAX_colli/2;j++{
			//放入第一张表中
			// 表一的该行没有满，接下来是放入表一，还没有达到最大的碰撞次数
		    if (table1_num[chart1]<table_bread && next_table1 == true && temp_colli <= MAX_colli){
				// 表1中 该行没有满
				if table1[chart1][table1_num[chart1]] =="" {
					table1[chart1][table1_num[chart1]] = cert1String
					if table1_num[chart1]< 4 {
						table1_num[chart1]++
					}
					next_table1 = true
					temp_colli = 0
					break
				// 解决因删除问题造成的空洞
				}else{
					// 找到空洞的位置
					for  table1[chart1][insert_time]!=""{
						insert_time++
					}
					if table1[chart1][insert_time]!="" {
						 fmt.Println("插入失败，请查询原因")
						for j := 0; j < 100; j++ {
							 fmt.Println("插入错误，请检查")
						}
						break
					}else{
						// 找到了空洞的位置
						table1[chart1][insert_time] = cert1String
						if table1_num[chart1]< 100 {
							table1_num[chart1]++
						}
						next_table1 = true
						temp_colli = 0
						insert_time = 0
						break
					}
				}
			// 表1满了，先放入表一，再碰撞放入表2，还未达到最大碰撞次数
			}else if (table1_num[chart1] == table_bread && next_table1 == true && temp_colli <= MAX_colli){
				// 第一张表满了，放入第二章表中
				//r=rand.New(rand.NewSource(time.Now().UnixNano()))
				// 随机找个地方换
				randplace := RandInt(4)
				// randplace := rand.Intn(100)
				//表1中的原数据
				temp_data = table1[chart1][randplace]
				// 错误验证
				if len(temp_data)==0 {
					 fmt.Println("此处数据为空")
					 fmt.Println(randplace)
				}
				fmt.Printf("\n")
				fmt.Println(temp_data)
				//将数据插入表中
				table1[chart1][randplace] = cert1String
				//放置临时数据，更新插入位置，插入表，插入数据，碰撞次数
				chart2 = compu_chart(temp_data,2)
				next_table1 = false
				cert1String = temp_data
				chart2 = compu_chart(temp_data,2)
				temp_colli++
				total_colli++
			// 开始放入第二表中，首先第二章表未满，轮到第二张表；未达到最大碰撞次数
			}else if (table2_num[chart2]<table_bread && next_table1 == false && temp_colli <= MAX_colli){
				// 第二章表为空
				if table2[chart2][table2_num[chart2]] =="" {
					table2[chart2][table2_num[chart2]] = cert1String
					if table2_num[chart2]< 4 {
						table2_num[chart2]++
					}
					next_table1 = true
					temp_colli = 0
					break
				}else{
					for  table2[chart2][insert_time] !=""{
						insert_time++						
					}
					if table2[chart2][insert_time] !="" {
						fmt.Println("插入错误，请检查")
						for j := 0; j < 100; j++ {
							fmt.Println("插入错误，请检查")
						}
						break
					}else{
						table2[chart2][insert_time] = cert1String
						if table2_num[chart2]< 100 {
							table2_num[chart2]++
						}
						next_table1 = true
						temp_colli = 0
						break
					}
				}
			// 开始放入第二表中，首先第二章表已满，轮到第二张表；未达到最大碰撞次数
			}else if (table2_num[chart2] ==table_bread && next_table1 == false && temp_colli <= MAX_colli) {
				// r =rand.New(rand.NewSource(time.Now().UnixNano()))
				// 在表中随机找个位置
				randplace := RandInt(4)
				// randplace := rand.Intn(100)
				// 临时的数据
				temp_data = table2[chart2][randplace]
				// 插入原数据
				table2[chart2][randplace] = cert1String
				// 准备插入临时数据
				chart1 = compu_chart(temp_data,1)
				next_table1 = true
				cert1String = temp_data
				temp_colli++
				total_colli++
			}else{
				MAX_colli_time++
				break
			}
		}
		
	}
	// 查询速度
	right_time:=0
	start := time.Now()
	for k:=20000;k<=1020000;k++ {
		
		fmt.Printf("第%d次开始查询",k-19999)
		fmt.Printf("---")
		//随机生成一个证书编号，是数字加随机字符串
		// cert1:=RandInt(1000000)
		//cert1String := strconv.Itoa(cert1) + RandString2(6)
		cert1String := strconv.Itoa(k*k)
		fmt.Printf(cert1String)
		fmt.Printf("---")
		
		//计算随机哈希，转换成int类型
		// cert2Hash :=MD5Hash2(cert1String)		
		// cert2String := cert2Hash[0:6]
		// cert2 := Hex2Dec2(cert2String)


		// 计算放入表的位置，一共有两个，一个是chart1,一个是chart2
		chart1 := compu_chart(cert1String,1)
		// 放入第二章表的数据
		chart2 := compu_chart(cert1String,2)
		cert1String = strconv.Itoa(compu_chart(cert1String,3))
		fmt.Printf("%d",chart1)
		fmt.Printf("---")
		fmt.Printf("%d",chart2)
		fmt.Printf("---")
		fmt.Printf("\n")
		
		for j:=0;j<4;j++{
			if table1[chart1][j]==cert1String||table2[chart2][j]==cert1String {
				right_time++
			}
		}
	}

	elapsed := time.Since(start).Nanoseconds()
	// 输出表的数据
	mm:=0
	nn:=0
	for m := 0; m < 2500; m++ {
		fmt.Printf("打印表1%d",table1_num[m])
		fmt.Printf("--")
		if table1_num[m]==4 {
			mm++
		}
	}
	for n := 0; n < 2500; n++ {
		fmt.Printf("打印表2%d",table2_num[n])
		fmt.Printf("--")
		if table2_num[n]==4 {
			nn++
		}
	}
	fmt.Printf("\n")
	// 表示已经满了的数据
	fmt.Println(mm)
	fmt.Println(nn)
	fmt.Printf("假阳性的次数%d\n",right_time)
	fmt.Printf("发生总共碰撞的此时为%d\n",total_colli)
	fmt.Printf("代码运行的总时间为%d\n",elapsed)

	
}
// 计算要放入的位置
func compu_chart(cert1String string,num int)int{
	// cert1,_ := strconv.Atoi(cert1String[0:len(cert1String)-6])
	cert1,_ := strconv.Atoi(cert1String)
	cert2Hash :=MD5Hash2(cert1String)
	fmt.Println("需要关注的地方")
	fmt.Println(cert2Hash)
	cert2String := cert2Hash[0:8]
	fmt.Println(cert2String)
	//指纹信息
	fp := Hex2Dec2(cert2String)%16384//函数将十六进制数转换为十进制数。
	if num == 1 {
		chart := cert1%5000
		return chart
	}
	if num == 2{
		chart := cert1%5000
		chart2 := (fp^chart)%5000
		// fmt.Println(chart)
		// fmt.Println(chart2)
		// fmt.Println(fp)
		// fmt.Println(chart2^fp)
		// fmt.Println("关注已结束")
		return chart2
	}
	if num==3{
		return fp
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
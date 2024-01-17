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
	MAX_colli:= 36  
	//最大的碰撞时间  
	MAX_colli_time := 0 
	//一次插入总共的碰撞次数
	total_colli := 0
	//临时的一次碰撞
	temp_colli := 0
	//表的长度
	//table_len := 100
	//表的宽度
	table_breadth := 16
	//临时的数据变量
	var temp_data string
	//插入次数计算
	insert_time := 0
	//表1的结构
	var table [4096][16]string
	//表一的附属表
	var table_num [4096]int

	//第n次的插入
	times := 0
	del_breadth := 0
	del_data := 0
//三组实验 95-124459-62259-62200  90-121239-58982-58980 85-117959-55705-55700
	//开始装填证书信息
	start := time.Now()
	for i:=1;i<=117959;i++ {
		if i==55705 {
			start = time.Now()
		}
		if i>=55700{
			del_breadth = RandInt(4096)
			del_data = RandInt(table_num[del_breadth])
			for table[del_breadth][del_data]==""{
				del_breadth = RandInt(4096)
				del_data = RandInt(table_num[del_breadth])
			}
			table[del_breadth][del_data]=""
			table_num[del_breadth]--
		}	 
			// if table_num[del_breadth]!=0{
			// 	fmt.Printf("在第%d行删除了第%d个格子，删除前为%s,\n",del_breadth,del_data,table[del_breadth][del_data])
			// 	fmt.Printf("删除后为%s,剩余了%d个指纹\n",table[del_breadth][del_data],table_num[del_breadth])
			// }else if table_num[del_breadth]==0{
			// 	del_breadth = RandInt(5000) 
			// 	del_data = RandInt(table_num[del_breadth])
			// 	fmt.Printf("在第%d行删除了第%d个格子，删除前为%s,\n",del_breadth,del_data,table[del_breadth][del_data])
			// 	table[del_breadth][del_data]=""
			// 	table_num[del_breadth]--
			// 	fmt.Printf("删除后为%s,剩余了%d个指纹\n",table[del_breadth][del_data],table_num[del_breadth])
			// }
		


		times++
		fmt.Printf("开始第%d次插入",times)
		fmt.Printf("---")
		row,_ := strconv.Atoi(compu_chart(strconv.Itoa(i),1))
		// chart2,_ := strconv.Atoi(compu_chart(cert1String,2))
		fpString := compu_chart(strconv.Itoa(i),3)
		fmt.Printf("第一个位置是%d",row)
		// fmt.Printf("---")
		// fmt.Printf("第二个位置是%d",chart2)
		fmt.Printf("---")
		fmt.Printf("指纹信息是%s",fpString)
		fmt.Printf("\n")
		// 哈希轮替考虑是否错误
		
		temp_colli = 0
		for j:=0;j<=MAX_colli;j++{
		    if (table_num[row]<table_breadth  && temp_colli <= MAX_colli){
				if table[row][table_num[row]] =="" {
					table[row][table_num[row]] = fpString
					table_num[row]++
					// fmt.Printf("在第%d行插入的指纹为%s,消耗的空格为%d\n",row,fpString,table_num[row])
					// for i := 0; i < 16; i++ {
					// 	fmt.Println(table[row][i])
					// }
					temp_colli = 0
					break
				}else{
					insert_time=0
					// fmt.Printf("测试的行数是%d,其中指纹有%s,\n",row,fpString)
					// for i := 0; i < 16; i++ {
					// 	fmt.Println(table[row][i])
					// }
					for  table[row][insert_time]!=""{
						insert_time++
					}
					if table[row][insert_time]!="" {
						 fmt.Println("插入失败，请查询原因")
						for k := 0; k < 100; k++ {
							 fmt.Println("插入错误，请检查")
						}
						// table1_num[chart1] = table_breadth
						break
					}else{
						table[row][insert_time] = fpString
						table_num[row]++
						// fmt.Printf("在第%d行插入的指纹为%s,消耗的空格为%d\n",row,fpString,table_num[row])
						// for i := 0; i < 16; i++ {
						// 	fmt.Println(table[row][i])
						// }
						temp_colli = 0
						break
					}
				}
			}else if(table_num[row] == table_breadth  && temp_colli <= MAX_colli){
				randplace := RandInt(16)
				temp_data = table[row][randplace]
				table[row][randplace] = fpString
				fpString = temp_data
				row = fptochart(fpString,row)
				temp_colli++
				total_colli++
			}
			// else if (table1_num[chart2]<table_breadth && next_hash1 == false && temp_colli <= MAX_colli){
			// 	if table1[chart2][table1_num[chart2]] =="" {
			// 		table1[chart2][table1_num[chart2]] = cert1String
			// 		table1_num[chart2]++
			// 		next_hash1 = true
			// 		temp_colli = 0
			// 		break
			// 	}else{
			// 		insert_time=0
			// 		for  table1[chart2][insert_time] !=""{
			// 			insert_time++	
			// 			if insert_time== table_breadth{
			// 				insert_time--
			// 				break
			// 			}					
			// 		}
			// 		if table1[chart2][insert_time] !="" {
			// 			for j := 0; j < 100; j++ {
			// 				fmt.Println("插入错误，请检查")
			// 			}
			// 			table1_num[chart2] = table_breadth
			// 			break
			// 		}else{
			// 			table1[chart2][insert_time] = cert1String
			// 			table1_num[chart2]++
			// 			next_hash1 = true
			// 			temp_colli = 0
			// 			break
			// 		}
			// 	}
			// }
			// else if (table1_num[chart2] ==table_breadth && next_hash1 == false && temp_colli <= MAX_colli) {
			// 	randplace := RandInt(16)
			// 	temp_data = table1[chart2][randplace]
			// 	table1[chart2][randplace] = cert1String
			// 	cert1String = temp_data
			// 	// 此处要修改指纹的获得方式
			// 	// 此处要修改指纹的获得方式
			// 	chart1 = fptochart(cert1String,1)
			// 	next_hash1 = true
			// 	temp_colli++
			// 	total_colli++
			// }
			if temp_colli == MAX_colli{
				MAX_colli_time++
				break
			}
		}
	}
	elapsed := time.Since(start).Nanoseconds()
	// right_num:=0
    // //假阳性测试
	// for j:=20000;j<1020000;j++{
	// 	string3:=strconv.Itoa(j*j)
	// 	fmt.Printf("查询次数为%d\n",j-19999)
	// 	chart1 := compu_chart(string3,1)
	// 	// 放入第二章表的数据
	// 	chart2 := compu_chart(string3,2)
	// 	for k:=0;k<4;k++{
	// 		if table1[chart1][k]==string3||table1[chart2][k]==string3 {
	// 			right_num++
	// 		}
	// 	}
	// }

	// right_num:=0
    //查询的速度
	// for j:=1;j<=76000;j++{
	// 	string3:=strconv.Itoa(j)
	// 	right_true:=false
	// 	fmt.Printf("查询次数为%d\n",j)
	// 	chart1,_ := strconv.Atoi(compu_chart(string3,1))
	// 	// 放入第二章表的数据
	// 	chart2,_ := strconv.Atoi(compu_chart(string3,2))
	// 	string3 = compu_chart(string3,3)
	// 	for k:=0;k<16;k++{
	// 		if table1[chart1][k]==string3||table1[chart2][k]==string3 {
	// 			right_num++
	// 			right_true = true
	// 		}
	// 		}
	// 	if right_true == false{
	// 		fmt.Printf("%d在%d和%d查询不到%s\n",j,chart1,chart2,string3)
	// 	}
	// }

	
	// 输出表的数据
	// mm:=0
	// // nn:=0
	// for m := 0; m < 5000; m++ {
	// 	fmt.Printf("这是第%d行",m)
	// 	if table_num[m]<=16 {
	// 		fmt.Printf("++")
	// 		fmt.Printf("%s",table[m][0])
	// 		fmt.Printf("++")
	// 		fmt.Printf("%s",table[m][1])
	// 		fmt.Printf("++")
	// 		fmt.Printf("%s",table[m][2])
	// 		fmt.Printf("++")
	// 		fmt.Printf("%s",table[m][3])
	// 		fmt.Printf("\n")
	// 		mm++
	// 	}
	// }
	// for n := 0; n < 2500; n++ {
	// 	fmt.Printf("%d",table2_num[n])
	// 	fmt.Printf("--")
	// 	if table2_num[n]==4 {
	// 		nn++
	// 	}
	// }
	fmt.Printf("\n")
	// 表示已经满了的数据
	//fmt.Println(mm)
	// fmt.Printf("假阳性概率为%d\n",right_num)
	// fmt.Println(nn)
	// fmt.Printf("查询成功的次数为%d\n",right_num)
	fmt.Println("75000-76000(95%)\n")
	fmt.Printf("插入失败的次数为%d\n",MAX_colli_time)
	fmt.Printf("发生总共碰撞的此时为%d\n",total_colli)
	fmt.Printf("代码运行的总时间为%d\n",elapsed)
}

// 计算要放入的位置
func compu_chart(cert1String string,num int)string{
	cert2Hash :=MD5Hash2(cert1String)
	// hash截取
	cert3String := cert2Hash[2:10]
	// 转换成十进制
	chart:=Hex2Dec2(cert3String)%4096
	// 指纹信息
	cert2String := cert2Hash[0:8]
	fp := (Hex2Dec2(cert2String))%16384
	chart2 := fp^chart
	if num == 1 {
		return strconv.Itoa(chart)
	}
	if num == 2{
		return strconv.Itoa(chart2%4096)
	}
	return strconv.Itoa(fp)+"--"+strconv.Itoa(chart)+"--"+strconv.Itoa(chart2)
}

func fptochart(fp string,chart int)int  {
	certinfo:=strings.Split(fp,"--")
	info1,_:=strconv.Atoi(certinfo[1])
	info2,_:=strconv.Atoi(certinfo[2])
	if chart == info1 {
		return info2%4096
	}
	return info1%4096
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

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
	insert_time := 0

	var table1 [100][100]string
	var table1_num [100]int
	var table2 [100][100]string
	var table2_num [100]int
	time1 := 0
	sun_time := 0

	//开始装填证书信息
	start := time.Now()
	for i:=0;i<=9000;i++ {
		if i>=8500 {
			sun_time = deleteOnedata(table1,table2,table1_num,table2_num)+sun_time
		}
		time1++
		fmt.Printf("%d",time1)
		fmt.Printf("---")
		//随机生成一个证书编号，是数字加随机字符串
		cert1:=RandInt(1000000)
		cert1String := strconv.Itoa(cert1) + RandString2(6)
		fmt.Printf(cert1String)
		fmt.Printf("---")
		
		//计算随机哈希，转换成int类型
		cert2Hash :=MD5Hash2(cert1String)		
		cert2String := cert2Hash[0:6]
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
			fmt.Println("zanting111")
			//放入第一张表中
		    if (table1_num[chart1]<table_bread && next_table1 == true && temp_colli <= MAX_colli){
				if table1[chart1][table1_num[chart1]] =="" {
					table1[chart1][table1_num[chart1]] = cert1String
					if table1_num[chart1]< 100 {
						table1_num[chart1]++
					}
					next_table1 = true
					temp_colli = 0
					fmt.Println("zanting222111")
					break
				}else{
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
						table1[chart1][insert_time] = cert1String
						if table1_num[chart1]< 100 {
							table1_num[chart1]++
						}
						next_table1 = true
						temp_colli = 0
						insert_time = 0
						break
					}
					fmt.Println("zanting222222")
				}
			}else if (table1_num[chart1] == table_bread && next_table1 == true && temp_colli <= MAX_colli){
				// 第一张表满了，放入第二章表中
				//r=rand.New(rand.NewSource(time.Now().UnixNano()))
				randplace := RandInt(100)
				// randplace := rand.Intn(100)
				temp_data = table1[chart1][randplace]
				if len(temp_data)==0 {
					fmt.Println("此处数据为空")
					fmt.Println(randplace)
				}
				fmt.Printf("\n")
				fmt.Printf("临时数据%d",temp_data)
				table1[chart1][randplace] = cert1String
				chart2 = compu_chart(temp_data,2)
				next_table1 = false
				cert1String = temp_data
				chart2 = compu_chart(temp_data,2)
				temp_colli++
				total_colli++
				fmt.Println("zanting333")
			}else if (table2_num[chart2]<table_bread && next_table1 == false && temp_colli <= MAX_colli){
				// 放入第二章表中
				if table2[chart2][table2_num[chart2]] =="" {
					table2[chart2][table2_num[chart2]] = cert1String
					if table2_num[chart2]< 100 {
						table2_num[chart2]++
					}
					next_table1 = true
					temp_colli = 0
					fmt.Println("zanting444")
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
						table2[chart2][insert_time] = cert2String
						if table2_num[chart2]< 100 {
							table2_num[chart2]++
						}
						next_table1 = true
						temp_colli = 0
						break
					}
					fmt.Println("zanting444")
				}
			}else if (table2_num[chart2] ==table_bread && next_table1 == false && temp_colli <= MAX_colli) {
				// r =rand.New(rand.NewSource(time.Now().UnixNano()))
				randplace := RandInt(100)
				// randplace := rand.Intn(100)
				temp_data = table2[chart2][randplace]
				table2[chart2][randplace] = cert1String
				chart1 = compu_chart(temp_data,1)
				next_table1 = true
				cert1String = temp_data
				temp_colli++
				total_colli++
				fmt.Println("zanting555")
			}else{
				MAX_colli_time++
				fmt.Println("zanting666")
				break
			}
		}
		
	}
	elapsed := time.Since(start).Nanoseconds()
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
	fmt.Printf("表1满的格数是：%d\n",mm)
	fmt.Printf("表2满的格数s是：%d\n",nn)
	fmt.Printf("发生总共碰撞的此时为%d\n",total_colli)
	fmt.Printf("代码运行的总时间为%d\n",elapsed/1000000000)
	fmt.Printf("发生假阳性的情况此时为%d\n",sun_time)
	
}
// 计算要放入的位置
func compu_chart(cert1String string,num int)int{
	cert1,_ := strconv.Atoi(cert1String[0:len(cert1String)-6])
	cert2Hash :=MD5Hash2(cert1String)
	cert2String := cert2Hash[0:6]
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

func deleteOnedata(table1 [100][100]string,table2 [100][100]string,table1_num [100]int,table2_num [100]int) int {
	del_sign := 1
	del_colli:=0
	k := 0
	m := 0 
	fmt.Printf("****1")
	for  del_sign ==1{
		fmt.Printf("****2")
		del1 := RandInt(100)
		del2 := RandInt(100)
		// 查重，看有没有同样的证书
		del_data:=table1[del1][del2]
		for del_data=="" {
			fmt.Printf("****3")
			del1 = RandInt(100)
			del2 = RandInt(100)
			// 查重，看有没有同样的证书
			del_data=table1[del1][del2]
		}
		del_chart1 := compu_chart(del_data,1)
		del_chart2 := compu_chart(del_data,2)
		del_time:=0
		// 终于找到错误地方了
		for del_time<table1_num[del_chart1] {
			fmt.Printf("****4")
			if table1[del_chart1][del_time]==del_data{
				k++
				if k>1 {
					fmt.Printf("发生了一次假阳性碰撞")
					del_colli++
					break
				}
			}
			del_time++
		}
		del_time =0
		for del_time<table2_num[del_chart2] {
			fmt.Printf("****5")
			if table2[del_chart2][del_time]==del_data{
				m++
				if k>1 {
					fmt.Printf("发生了一次假阳性碰撞")
					del_colli++
					break
				}
			}
			del_time++
		}


		if del1%2 ==0{
			if table1[del1][del2]!="" {
				table1[del1][del2] = ""
				del_sign =0
				table1_num[del1]--
			}	
		}else{
			if table2[del1][del2]!="" {
				table2[del1][del2] = ""
				del_sign=0
				table2_num[del1]--
			}
		}
	}
	return (k+m)		
}


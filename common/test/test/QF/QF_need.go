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
// 商过滤器代码
func main () {
	//开始设计一个表格
	// 定义数据
	//一共有多少槽
	MAX_slot:= 100000
	//每个槽中有3个bit 
	MAX_yushu := 16

	q:=5
	r:=4

	//表1的结构
	var table1 [100000][16]int


	//开始装填证书信息
	for i:=0;i<=19000;i++ {

		//开始第一次插入
		time1++
		fmt.Printf("time1是%d次插入",time1)
		fmt.Printf("---")

		shangstring,yushustring := compu_chart(i*i)
		shangint,_ := strconv.Atoi(shangstring)
		yushuint,_ := strconv.Atoi(yushustring)

		insert := false
		//开始插入表中		
        for insert==false{
            //直接插入
            if table1[shangint][0]==0 {
					table1[cert1int][0] = yushuint
					insert = true
				}
			MAX_yushu:=0
			}else{
				i := 0
				for table1[shangint][i]!=0 {
					if(table1[shangint][i]>yushuint){
						MAX_yushu = table1[shangint][i]
						table1[shangint][i] = yushuint
						yushuint = MAX_yushu
						i++
					}else(table1[shangint][i]<yushuint){
						i++
					}else{
						insert = true
						break
					}
				}
			}

			//time is add
			times:=shangint
			if table1[times+1][0]!= 0 {
				for i := 0; i < 16; i++ {
					table1[shangint][i]=table1[shangint][i]+1
					table1[shangint][i]=table1[shangint][i]-1
				}
				times=times+1
			}
			
			
% 放入第二章表的数据
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
func compu_chart(cert1String string)(string,string){
	// cert1,_ := strconv.Atoi(cert1String[0:len(cert1String)-6])
	cert1,_ := strconv.Atoi(cert1String)
	cert2Hash :=MD5Hash2(cert1String)
	fmt.Println("需要关注的地方")
	fmt.Println(cert2Hash)
	cert1String := cert2Hash[0:q-1]
	fmt.Println(cert1String)
	cert2String := cert2Hash[(q-1):(q+r-1)]
	fmt.Println(cert2String)
    return cert1String,cert2String
}

func MD5Hash2(data string) string {
    t := md5.New()
    io.WriteString(t, data)
    return fmt.Sprintf("%x", t.Sum(nil))
}




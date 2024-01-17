package main

import "fmt"

func main () {
	var server_number int
	_,_ = fmt.scan(&server_number)
	server_data:=[server_number][5]int
	for i:=0;i< server_number;i++{
		_,_ = fmt.scan(&server_data[i][0],&server_data[i][1],&server_data[i][2],&server_data[i][3],&server_data[i][4])
	}
	request_data:=[6]int
	_,_ = fmt.scan(&request_data][0],&request_data][1],&request_data][2],&request_data][3],&request_data][4],&request_data][5],&request_data][6])
	
	//分配为1
	number:=[server_number]int
	total:= server_number
	total_number:=[server_number]int
	if request_data[1]==1{
       for j := 0; j < server_number; j++ {
		   if server_data[j][1]<request_data[2] {
			  number[j]==1
			  total--
		   }else if server_data[j][2]<request_data[3] {
			  number[j]==1
			  total--
		   }else if request_data[4]!=9 || server_data[j][3]!=request_data[4] {
			  number[j]==1
			  total--
		   }else request_data[5]!=9 || server_data[j][4]!=request_data[5]{
			 number[j]==1
			 total--
		   }	   
	   }
	}else{
		for j := 0; j < server_number; j++ {
			if server_data[j][2]<request_data[3] {
			   number[j]==1
			   total--
			}else if server_data[j][1]<request_data[2] {
			   number[j]==1
			   total--
			}else if request_data[4]!=9 || server_data[j][3]!=request_data[4] {
			   number[j]==1
			   total--
			}else request_data[5]!=9 || server_data[j][4]!=request_data[5]{
			  number[j]==1
			  total--
			}	   
		}
	}



	temp:=0
	for k:=0;k< server_number;k++{
		if number[k]==0{
			total_number[temp]
			temp++
		}
	}
	fmt.Println(total,total_number[1],total_number[2],total_number[3])
}

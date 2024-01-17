package test

func createCertificate(apltp, info, ts, pk, sig) Hcert {
	//申请人信息验证
    if true == verify(info, pk, sig)  then
	//申请人申请类型验证
        if apltp == iss | apltp == ren
	//颁发新的证书
            cert=signcert(info,pk)
	//证书上链
            send cert == IC
	//获取链上信息
	        if cert onchaining==true then
			    //构造Hcert
			    Hcert =signHcert(cert,CA_id,block_number,tx_id)
				//返回Hcert
				return Hcert
		    else return nil
		else return nil
	else return nil	
	
}

func certificaterevoke(apltp, info, Hcert, ts, pk, sig) bool {
	//申请人信息验证
    if true == verify(info, pk, sig)  then
	//申请人申请类型验证
	    if apltp == rev then
		    //撤销证书公钥上链
			cert,CA_id,block_number,tx_id =Extract(Hcert)
			if cert == HCV(CA_id,block_number,tx_id) then
			   //撤销证书插入MCF
			   fp=SHA256(cert)
			   if true==MCF_insert(fp) then
			   return true
			else return false
		else return false
	else return false
}

func certificateauth(apltp, info, Hcert, pk, sig) bool {
	//申请人信息验证
    if true == verify(info, pk, sig)  then   
	//申请人申请类型验证
	    cert,CA_id,block_number,tx_id =Extract(Hcert)
	    if cert == HCV(CA_id,block_number,tx_id) then
		   fp=SHA256(cert)
		   if true==search(fp) then
		      true==traversal_chain(cert) then
			  return false //证书过期
			else return true  //证书认证通过
		else return true
	else return true
			

	//链上证书查询
	//MCF证书验证
	//公钥验证
}
/*CA迁移，从一条链迁移到另外一条链，需要oldcert，newcert,CA_new,分布式密钥，密钥更新，
*验证Ca，
*
*/
func CAmirgation(Hcert_{old},cert_{new}, pk, sig, {s1,s2...sk}) bool {
	// IC链上变色龙哈希修改
	if true == verify(cert_{old}, pk, sig)  then
	   //密钥收集
       if true == Verify_{s}(s1,s2...sk) then
	      s = recon(s1,s2...sk)
		  cert,CA_id,block_number,tx_id =Extract(Hcert_{old})
		  // 撤销链上证书删除
		  bool = chameleonHash(s,cert,r,cert_{new})
	      send chameleonHash == SC
		  if bool == true then
		     send cert_{old} == RC
			 return true
		  else 
		    bool = chameleonHash(s,cert_{new},r,cert)
		    return false
}

	
~\\
$\textbf{On}$ Expired$\_$Cert(tx):\\
\ \ \ \ \		cert $\gets$ Tx$\_$analysis(tx)\\
\ \ \ \ \		HCV$\_$verify(cert)\\
\ \ \ \ \		Revoke$\_$chain$\_$verify(cert)\\
\ \ \ \ \		Delete$\_$fingerprint$\_$from$\_$${\mbox{DCF}_1}$(cert's ${\mbox{fingerprint}_1}$) \\
\ \ \ \ \		Delete$\_$fingerprint$\_$from$\_$${\mbox{DCF}_2}$(cert's ${\mbox{fingerprint}_2}$) \\
\ \ \ \ \		$\textbf{return}$ result\\

~\\
$\textbf{On}$ Audit$\_$Cert(tx):\\
\ \ \ \ \		cert $\gets$ Tx$\_$analysis(tx)\\
\ \ \ \ \	HCV$\_$verify(cert)\\
\ \ \ \ \		fingerprint$\_$verify$\_$from$\_$${\mbox{DCF}_1}$(cert's ${\mbox{fingerprint}_1}$) \\
\ \ \ \ \		fingerprint$\_$verify$\_$from$\_$${\mbox{DCF}_2}$(cert's ${\mbox{fingerprint}_2}$) \\
\ \ \ \ \		$\textbf{return}$ result\\
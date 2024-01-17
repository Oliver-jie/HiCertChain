/****************************************************** 
 *  Copyright 2018 IBM Corporation 
 *  Licensed under the Apache License, Version 2.0 (the "License"); 
 *  you may not use this file except in compliance with the License. 
 *  You may obtain a copy of the License at 
 *  http://www.apache.org/licenses/LICENSE-2.0 
 *  Unless required by applicable law or agreed to in writing, software 
 *  distributed under the License is distributed on an "AS IS" BASIS, 
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. 
 *  See the License for the specific language governing permissions and 
 *  limitations under the License.
 */
package org.example.network;

import java.io.File;
import java.io.FileInputStream;
import java.io.IOException;
import java.lang.reflect.InvocationTargetException;
import java.security.NoSuchAlgorithmException;
import java.security.spec.InvalidKeySpecException;
import java.util.*;
import java.util.logging.Level;
import java.util.logging.Logger;

import org.apache.commons.compress.utils.IOUtils;
import org.example.chaincode.invocation.InvokeChaincode;
import org.example.client.CAClient;
import org.example.client.FabricClient;
import org.example.config.Config;
import org.example.user.UserContext;
import org.example.util.FileUtils;
import org.example.util.Util;
import org.hyperledger.fabric.sdk.*;
import org.hyperledger.fabric.sdk.exception.CryptoException;
import org.hyperledger.fabric.sdk.exception.ProposalException;
import org.hyperledger.fabric.sdk.exception.TransactionException;
import org.hyperledger.fabric.sdk.security.CryptoSuite;
import org.hyperledger.fabric_ca.sdk.exception.InvalidArgumentException;

/**
 * 
 * @author Balaji Kadambi
 *
 */

public class CreateChannel {

	public static void proseccs() {
		try {
			CryptoSuite.Factory.getCryptoSuite();
			Util.cleanUp();
			// Construct Channel
			UserContext adminUser = new UserContext();
			adminUser.setName(Config.ADMIN);
			adminUser.setAffiliation(Config.ORG2);
			adminUser.setMspId(Config.ORG2_MSP);
			File f = new File (InvokeChaincode.class.getResource("/ca.crt").getPath());
			String certficate = new String (IOUtils.toByteArray(new FileInputStream(f)),"UTF-8");
			Properties properties = new Properties();
			properties.put("pemBytes", certficate.getBytes());
			properties.setProperty("allowAllHostNames", "true");
			CAClient caclient=new  CAClient(Config.CA_ORG2_URL, properties);
			caclient.setAdminUserContext(adminUser);
			adminUser =  caclient.enrollAdminUserTLS("admin", "adminpw");

			UserContext adminUser1 = new UserContext();
			adminUser1.setName(Config.ADMIN);
			adminUser1.setAffiliation(Config.ORG1);
			adminUser1.setMspId(Config.ORG1_MSP);
			File f1 = new File (InvokeChaincode.class.getResource("/ca1.crt").getPath());
			String certficate1 = new String (IOUtils.toByteArray(new FileInputStream(f1)),"UTF-8");
			Properties properties1 = new Properties();
			properties1.put("pemBytes", certficate1.getBytes());
			properties1.setProperty("allowAllHostNames", "true");
			CAClient caclient1=new  CAClient(Config.CA_ORG1_URL, properties1);
			caclient1.setAdminUserContext(adminUser1);
			adminUser1 =  caclient1.enrollAdminUserTLS("admin", "adminpw");



			/*UserContext org1Admin = new UserContext();
			Enrollment enrollOrg1Admin = Util.getEnrollment("C:\\Users\\13202\\Desktop\\linux-fabric\\blockchain-application-using-fabric-java-sdk\\java\\src\\main\\resources\\key1\\org1_key", null,
					"C:\\Users\\13202\\Desktop\\linux-fabric\\blockchain-application-using-fabric-java-sdk\\java\\src\\main\\resources\\cert1\\org1_cert", null);
			org1Admin.setEnrollment(enrollOrg1Admin);
			org1Admin.setMspId(Config.ORG1_MSP);
			org1Admin.setName(Config.ADMIN);

			UserContext org2Admin = new UserContext();
			Enrollment enrollOrg2Admin = Util.getEnrollment("C:\\Users\\13202\\Desktop\\linux-fabric\\blockchain-application-using-fabric-java-sdk\\java\\src\\main\\resources\\key2\\org2_key", null,
					"C:\\Users\\13202\\Desktop\\linux-fabric\\blockchain-application-using-fabric-java-sdk\\java\\src\\main\\resources\\cert2\\org2_cert", null);
			org2Admin.setEnrollment(enrollOrg2Admin);
			org2Admin.setMspId(Config.ORG2_MSP);
			org2Admin.setName(Config.ADMIN);*/

			FabricClient fabClient = new FabricClient(adminUser1);

			// Create a new channel
            File fpOder = new File (InvokeChaincode.class.getResource("/order.crt").getPath());
            String certficatepOder = new String (IOUtils.toByteArray(new FileInputStream(fpOder)),"UTF-8");
            Properties peer_propertiesOder = new Properties();
            peer_propertiesOder.put("pemBytes", certficatepOder.getBytes());
            peer_propertiesOder.setProperty("sslProvider", "openSSL");
            peer_propertiesOder.setProperty("negotiationType", "TLS");
            Orderer orderer = fabClient.getInstance().newOrderer(Config.ORDERER_NAME, Config.ORDERER_URL,peer_propertiesOder);
			//ChannelConfiguration channelConfiguration = new ChannelConfiguration(new File("C:\\Users\\13202\\Desktop\\linux-fabric\\blockchain-application-using-fabric-java-sdk\\java\\src\\main\\resources\\mychannel.tx"));

			//byte[] channelConfigurationSignatures = fabClient.getInstance()
				//	.getChannelConfigurationSignature(channelConfiguration, adminUser1);

			Channel mychannel = fabClient.getInstance().newChannel("channel");


            File fp = new File (InvokeChaincode.class.getResource("/server-org2p0.crt").getPath());
            String certficatep = new String (IOUtils.toByteArray(new FileInputStream(fp)),"UTF-8");
            Properties peer_properties = new Properties();
            peer_properties.put("pemBytes", certficatep.getBytes());
            peer_properties.setProperty("sslProvider", "openSSL");
            peer_properties.setProperty("negotiationType", "TLS");
            Peer peer0_org2 = fabClient.getInstance().newPeer(Config.ORG2_PEER_0, Config.ORG2_PEER_0_URL,peer_properties);
            //EventHub eventHub = fabClient.getInstance().newEventHub("eventhub01", "grpc://localhost:7053");


            File fp2 = new File (InvokeChaincode.class.getResource("/server-org1p0.crt").getPath());
            String certficatep2 = new String (IOUtils.toByteArray(new FileInputStream(fp2)),"UTF-8");
            Properties peer_properties2= new Properties();
            peer_properties2.put("pemBytes", certficatep2.getBytes());
            peer_properties2.setProperty("sslProvider", "openSSL");
            peer_properties2.setProperty("negotiationType", "TLS");
            Peer peer0_org1 = fabClient.getInstance().newPeer(Config.ORG1_PEER_0, Config.ORG1_PEER_0_URL,peer_properties2);
			mychannel.addOrderer(orderer);
			mychannel.joinPeer(peer0_org1);

			mychannel.initialize();
			
			fabClient.getInstance().setUserContext(adminUser);
			mychannel = fabClient.getInstance().getChannel("channel");
			mychannel.joinPeer(peer0_org2);
			
			Logger.getLogger(CreateChannel.class.getName()).log(Level.INFO, "Channel created "+mychannel.getName());
            Collection peers = mychannel.getPeers();
            Iterator peerIter = peers.iterator();
            while (peerIter.hasNext())
            {
            	  Peer pr = (Peer) peerIter.next();
            	  Logger.getLogger(CreateChannel.class.getName()).log(Level.INFO,pr.getName()+ " at " + pr.getUrl());
            }
            
		} catch (Exception e) {
			e.printStackTrace();
		}
	}
	// initialize Channel
	private static Channel initializeChannel(HFClient client, String channelName) throws InvalidArgumentException, TransactionException, org.hyperledger.fabric.sdk.exception.InvalidArgumentException, IOException, ProposalException, NoSuchAlgorithmException, CryptoException, InvalidKeySpecException, ClassNotFoundException, NoSuchMethodException, InvocationTargetException, InstantiationException, IllegalAccessException {
		Properties orderer1Prop = new Properties();
		orderer1Prop.setProperty("pemFile", "C:\\Users\\13202\\Desktop\\linux-fabric\\blockchain-application-using-fabric-java-sdk\\java\\src\\main\\resources\\order.crt");
		orderer1Prop.setProperty("sslProvider", "openSSL");
		orderer1Prop.setProperty("negotiationType", "TLS");
		orderer1Prop.setProperty("hostnameOverride", "orderer.example.com");
		orderer1Prop.setProperty("trustServerCertificate", "true");
		orderer1Prop.put("grpc.NettyChannelBuilderOption.maxInboundMessageSize", 9000000);
		Orderer orderer = client.newOrderer("orderer.example.com", "grpcs://orderer.example.com:7050", orderer1Prop);

		Properties peer1Prop = new Properties();
		peer1Prop.setProperty("pemFile", "C:\\Users\\13202\\Desktop\\linux-fabric\\blockchain-application-using-fabric-java-sdk\\java\\src\\main\\resources\\ca1.crt");
		peer1Prop.setProperty("sslProvider", "openSSL");
		peer1Prop.setProperty("negotiationType", "TLS");
		peer1Prop.setProperty("hostnameOverride", "peer0.org1.example.com");
		peer1Prop.setProperty("trustServerCertificate", "true");
		peer1Prop.put("grpc.NettyChannelBuilderOption.maxInboundMessageSize", 9000000);
		Peer peer = client.newPeer("peer0.org1.example.com", "grpcs://peer0.org1.example.com:7051", peer1Prop);

		Properties peer2Prop = new Properties();
		peer2Prop.setProperty("pemFile", "C:\\Users\\13202\\Desktop\\linux-fabric\\blockchain-application-using-fabric-java-sdk\\java\\src\\main\\resources\\ca.crt");
		peer2Prop.setProperty("sslProvider", "openSSL");
		peer2Prop.setProperty("negotiationType", "TLS");
		peer2Prop.setProperty("hostnameOverride", "peer0.org2.example.com");
		peer2Prop.setProperty("trustServerCertificate", "true");
		peer2Prop.put("grpc.NettyChannelBuilderOption.maxInboundMessageSize", 9000000);
		Peer peer2 = client.newPeer("peer0.org2.example.com", "grpcs://peer0.org2.example.com:9051", peer2Prop);
		//EventHub eventHub = client.newEventHub("peer1.operation.com",Config.EVENTHUB1_ADDRESS,peer1Prop);
		ChannelConfiguration channelConfiguration = new ChannelConfiguration(new File("C:\\Users\\13202\\Desktop\\linux-fabric\\blockchain-application-using-fabric-java-sdk\\java\\src\\main\\resources\\mychxx.tx"));
		//3. 获取签名
		byte[] signData = client.getChannelConfigurationSignature(channelConfiguration, client.getUserContext());
		Channel channel = client.newChannel(channelName,orderer,channelConfiguration,signData);
		channel.addOrderer(orderer);
		channel.initialize();
		channel.serializeChannel(new File("C:\\Users\\13202\\Desktop\\linux-fabric\\blockchain-application-using-fabric-java-sdk\\java\\src\\main\\resources\\block\\"+channelName + ".block"));
		String channel64String=Base64.getEncoder().encodeToString(channel.serializeChannel());
		FileUtils.writeFile(channel64String, "C:\\Users\\13202\\Desktop\\linux-fabric\\blockchain-application-using-fabric-java-sdk\\java\\src\\main\\resources\\channelString\\" + channelName + ".txt", false);
		System.out.println("join peer");
		channel.joinPeer(peer);
		UserContext org2Admin = new UserContext();
		Enrollment enrollOrg2Admin = Util.getEnrollment("C:\\Users\\13202\\Desktop\\linux-fabric\\blockchain-application-using-fabric-java-sdk\\java\\src\\main\\resources\\key2\\org2_key", null,
				"C:\\Users\\13202\\Desktop\\linux-fabric\\blockchain-application-using-fabric-java-sdk\\java\\src\\main\\resources\\server-org2p0.crt", null);
		org2Admin.setEnrollment(enrollOrg2Admin);
		org2Admin.setMspId(Config.ORG2_MSP);
		org2Admin.setName(Config.ADMIN);
		FabricClient fabClient = new FabricClient(org2Admin);
		String channel64String2 = new String(FileUtils.readFile("C:\\Users\\13202\\Desktop\\linux-fabric\\blockchain-application-using-fabric-java-sdk\\java\\src\\main\\resources\\channelString\\" + channelName + ".txt"));
		Channel channel1 = fabClient.getInstance().deSerializeChannel(Base64.getDecoder().decode(channel64String2.getBytes()));
		channel1.initialize();
		channel1.joinPeer(peer2);
		System.out.println("channel初始化");
		return channel;
	}

	public static void main(String[] args) throws Exception {
		CryptoSuite.Factory.getCryptoSuite();
		Util.cleanUp();
		// Construct Channel
		/*UserContext adminUser1 = new UserContext();
		adminUser1.setName(Config.ADMIN);
		adminUser1.setAffiliation(Config.ORG1);
		adminUser1.setMspId(Config.ORG1_MSP);
		File f1 = new File (InvokeChaincode.class.getResource("/ca1.crt").getPath());
		String certficate1 = new String (IOUtils.toByteArray(new FileInputStream(f1)),"UTF-8");
		Properties properties1 = new Properties();
		properties1.put("pemBytes", certficate1.getBytes());
		properties1.setProperty("allowAllHostNames", "true");*/
		//CAClient caclient1=new  CAClient(Config.CA_ORG1_URL, properties1);
		//caclient1.setAdminUserContext(adminUser1);
		//adminUser1 =  caclient1.enrollAdminUserTLS("admin", "adminpw");


		UserContext org1Admin = new UserContext();
		Enrollment enrollOrg1Admin = Util.getEnrollment("D:\\linux-fabric\\blockchain-application-using-fabric-java-sdk\\java\\src\\main\\resources\\key1\\org1_key", null,
				"C:\\Users\\13202\\Desktop\\linux-fabric\\blockchain-application-using-fabric-java-sdk\\java\\src\\main\\resources\\server-org1p0.crt", null);
		org1Admin.setEnrollment(enrollOrg1Admin);
		org1Admin.setMspId(Config.ORG1_MSP);
		org1Admin.setName(Config.ADMIN);
		FabricClient fabClient = new FabricClient(org1Admin);
		initializeChannel(fabClient.getInstance(),"mychxx");
	}

}

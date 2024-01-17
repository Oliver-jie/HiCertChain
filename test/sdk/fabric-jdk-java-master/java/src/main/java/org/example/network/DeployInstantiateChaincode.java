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
import java.nio.file.Paths;
import java.util.ArrayList;
import java.util.Collection;
import java.util.List;
import java.util.Properties;
import java.util.concurrent.CompletableFuture;
import java.util.logging.Level;
import java.util.logging.Logger;

import org.example.client.ChannelClient;
import org.example.client.FabricClient;
import org.example.config.Config;
import org.example.user.UserContext;
import org.example.util.Util;
import org.hyperledger.fabric.sdk.*;
import org.hyperledger.fabric.sdk.TransactionRequest.Type;
import org.hyperledger.fabric.sdk.exception.ProposalException;
import org.hyperledger.fabric.sdk.security.CryptoSuite;
import org.hyperledger.fabric_ca.sdk.exception.InvalidArgumentException;

/**
 * 
 * @author Balaji Kadambi
 *
 */

public class DeployInstantiateChaincode {

	public static void main(String[] args) {
		try {
			CryptoSuite cryptoSuite = CryptoSuite.Factory.getCryptoSuite();

			UserContext org1Admin = new UserContext();
			Enrollment enrollOrg1Admin = Util.getEnrollment("C:\\Users\\13202\\Desktop\\linux-fabric\\blockchain-application-using-fabric-java-sdk\\java\\src\\main\\resources\\key1\\org1_key", null,
					"C:\\Users\\13202\\Desktop\\linux-fabric\\blockchain-application-using-fabric-java-sdk\\java\\src\\main\\resources\\server-org1p0.crt", null);
			org1Admin.setEnrollment(enrollOrg1Admin);
			org1Admin.setMspId(Config.ORG1_MSP);
			org1Admin.setName(Config.ADMIN);
			
			UserContext org2Admin = new UserContext();
			Enrollment enrollOrg2Admin = Util.getEnrollment("C:\\Users\\13202\\Desktop\\linux-fabric\\blockchain-application-using-fabric-java-sdk\\java\\src\\main\\resources\\key2\\org2_key", null,
					"C:\\Users\\13202\\Desktop\\linux-fabric\\blockchain-application-using-fabric-java-sdk\\java\\src\\main\\resources\\server-org2p0.crt", null);
			org2Admin.setEnrollment(enrollOrg2Admin);
			org2Admin.setMspId(Config.ORG2_MSP);
			org2Admin.setName(Config.ADMIN);
			FabricClient fabClient = new FabricClient(org1Admin);
			HFClient client = fabClient.getInstance();
			Channel mychannel = fabClient.getInstance().newChannel("mychannel");
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
			mychannel.addOrderer(orderer);
			mychannel.addPeer(peer);
			mychannel.addPeer(peer2);
			mychannel.initialize();

			List<Peer> org1Peers = new ArrayList<Peer>();
			org1Peers.add(peer);
			
			List<Peer> org2Peers = new ArrayList<Peer>();
			org2Peers.add(peer2);

			String package1 = fabClient.deployChainCode(Config.CHAINCODE_1_NAME,
					Config.CHAINCODE_1_PATH, Config.CHAINCODE_ROOT_DIR, TransactionRequest.Type.GO_LANG,
					Config.CHAINCODE_1_VERSION, org1Peers);



			fabClient.getInstance().setUserContext(org2Admin);

			String package2 = fabClient.deployChainCode(Config.CHAINCODE_1_NAME,
					Config.CHAINCODE_1_PATH, Config.CHAINCODE_ROOT_DIR, TransactionRequest.Type.GO_LANG,
					Config.CHAINCODE_1_VERSION, org2Peers);



			ChannelClient channelClient = new ChannelClient(mychannel.getName(), mychannel, fabClient);

			String[] arguments = { "" };
			CompletableFuture<BlockEvent.TransactionEvent> init = channelClient.instantiateChainCode(Config.CHAINCODE_1_NAME, Config.CHAINCODE_1_VERSION,
					Config.CHAINCODE_1_PATH, TransactionRequest.Type.GO_LANG.toString(), "init", arguments, null, package1, org1Peers);


		} catch (Exception e) {
			e.printStackTrace();
		}
	}
}

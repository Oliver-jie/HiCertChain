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
package org.example.chaincode.invocation;

import static java.nio.charset.StandardCharsets.UTF_8;

import java.io.File;
import java.io.FileInputStream;
import java.util.Collection;
import java.util.HashMap;
import java.util.Map;
import java.util.Properties;
import java.util.logging.Level;
import java.util.logging.Logger;

import org.apache.commons.compress.utils.IOUtils;
import org.example.client.CAClient;
import org.example.client.ChannelClient;
import org.example.client.FabricClient;
import org.example.config.Config;
import org.example.user.UserContext;
import org.example.util.Util;
import org.hyperledger.fabric.sdk.ChaincodeID;
import org.hyperledger.fabric.sdk.ChaincodeResponse.Status;
import org.hyperledger.fabric.sdk.Channel;

import org.hyperledger.fabric.sdk.Orderer;
import org.hyperledger.fabric.sdk.Peer;
import org.hyperledger.fabric.sdk.ProposalResponse;
import org.hyperledger.fabric.sdk.TransactionProposalRequest;

/**
 * 
 * @author Balaji Kadambi
 *
 */

public class InvokeChaincode {

	private static final byte[] EXPECTED_EVENT_DATA = "!".getBytes(UTF_8);
	private static final String EXPECTED_EVENT_NAME = "event";

	public static void main(String args[]) {
		try {
            Util.cleanUp();
			UserContext adminUser = new UserContext();
			adminUser.setName(Config.ADMIN);
			adminUser.setAffiliation(Config.ORG2);
			adminUser.setMspId(Config.ORG2_MSP);
			File f = new File (InvokeChaincode.class.getResource("/ca.crt").getPath());
			String certficate = new String (IOUtils.toByteArray(new FileInputStream(f)),"UTF-8");
			Properties properties = new Properties();
			properties.put("pemBytes", certficate.getBytes());
			properties.setProperty("pemFile", f.getAbsolutePath());
			properties.setProperty("allowAllHostNames", "true");
			CAClient caclient=new  CAClient(Config.CA_ORG2_URL, properties);
			caclient.setAdminUserContext(adminUser);
			adminUser =  caclient.enrollAdminUserTLS("admin", "adminpw");
			FabricClient fabClient = new FabricClient(adminUser);
			ChannelClient channelClient = fabClient.createChannelClient(Config.CHANNEL_NAME);
			Channel channel = channelClient.getChannel();
			File fp = new File (InvokeChaincode.class.getResource("/ca.crt").getPath());
			String certficatep = new String (IOUtils.toByteArray(new FileInputStream(fp)),"UTF-8");
			Properties peer_properties = new Properties();
			peer_properties.put("pemBytes", certficatep.getBytes());
			peer_properties.setProperty("sslProvider", "openSSL");
			peer_properties.setProperty("negotiationType", "TLS");
			Peer peer = fabClient.getInstance().newPeer(Config.ORG2_PEER_0, Config.ORG2_PEER_0_URL,peer_properties);
			//EventHub eventHub = fabClient.getInstance().newEventHub("eventhub01", "grpc://localhost:7053");
			File fpOder = new File (InvokeChaincode.class.getResource("/order.crt").getPath());
			String certficatepOder = new String (IOUtils.toByteArray(new FileInputStream(fpOder)),"UTF-8");
			Properties peer_propertiesOder = new Properties();
			peer_propertiesOder.put("pemBytes", certficatepOder.getBytes());
			peer_propertiesOder.setProperty("sslProvider", "openSSL");
			peer_propertiesOder.setProperty("negotiationType", "TLS");
			Orderer orderer = fabClient.getInstance().newOrderer(Config.ORDERER_NAME, Config.ORDERER_URL,peer_propertiesOder);

			File fp2 = new File (InvokeChaincode.class.getResource("/ca1.crt").getPath());
			String certficatep2 = new String (IOUtils.toByteArray(new FileInputStream(fp2)),"UTF-8");
			Properties peer_properties2= new Properties();
			peer_properties2.put("pemBytes", certficatep2.getBytes());
			peer_properties2.setProperty("sslProvider", "openSSL");
			peer_properties2.setProperty("negotiationType", "TLS");
			Peer peer2 = fabClient.getInstance().newPeer(Config.ORG1_PEER_0, Config.ORG1_PEER_0_URL,peer_properties2);
			channel.addPeer(peer);
			channel.addPeer(peer2);
			//channel.addEventHub(eventHub);
			channel.addOrderer(orderer);
			channel.initialize();

			TransactionProposalRequest request = fabClient.getInstance().newTransactionProposalRequest();
			ChaincodeID ccid = ChaincodeID.newBuilder().setName("fabcar").build();
			request.setChaincodeID(ccid);
			request.setFcn("createCar");
			String[] arguments = { "CAR15", "Mini", "Volt", "Red", "pandaUUU" };
			request.setArgs(arguments);
			request.setProposalWaitTime(1000);

			Map<String, byte[]> tm2 = new HashMap<>();
			tm2.put("HyperLedgerFabric", "TransactionProposalRequest:JavaSDK".getBytes(UTF_8)); 																								
			tm2.put("method", "TransactionProposalRequest".getBytes(UTF_8)); 
			tm2.put("result", ":)".getBytes(UTF_8));
			tm2.put(EXPECTED_EVENT_NAME, EXPECTED_EVENT_DATA); 
			request.setTransientMap(tm2);
			Collection<ProposalResponse> responses = channelClient.sendTransactionProposal(request);
			for (ProposalResponse res: responses) {
				Status status = res.getStatus();
				Logger.getLogger(InvokeChaincode.class.getName()).log(Level.INFO,"Invoked createCar on "+Config.CHAINCODE_1_NAME + ". Status - " + status);
			}
									
		} catch (Exception e) {
			e.printStackTrace();
		}
	}

}

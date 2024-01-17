package org.example.controller;

import org.apache.commons.compress.utils.IOUtils;
import org.apache.commons.logging.LogFactory;
import org.example.chaincode.invocation.InvokeChaincode;
import org.example.chaincode.invocation.QueryChaincode;
import org.example.client.CAClient;
import org.example.client.ChannelClient;
import org.example.client.FabricClient;
import org.example.config.Config;
import org.example.network.DeployInstantiateChaincode;
import org.example.user.UserContext;
import org.example.util.Util;
import org.hyperledger.fabric.sdk.*;
import org.hyperledger.fabric.sdk.exception.CryptoException;
import org.hyperledger.fabric.sdk.exception.InvalidArgumentException;
import org.hyperledger.fabric.sdk.exception.ProposalException;
import org.hyperledger.fabric.sdk.exception.TransactionException;
import org.hyperledger.fabric.sdk.security.CryptoSuite;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.io.IOException;
import java.lang.reflect.InvocationTargetException;
import java.security.NoSuchAlgorithmException;
import java.security.spec.InvalidKeySpecException;
import java.util.*;
import java.util.concurrent.CompletableFuture;
import java.util.logging.Level;
import java.util.logging.Logger;

import static java.nio.charset.StandardCharsets.UTF_8;
import static org.hyperledger.fabric.sdk.Channel.PeerOptions.createPeerOptions;

@RestController
public class ContrllerTest {
    private static final byte[] EXPECTED_EVENT_DATA = "!".getBytes(UTF_8);
    private static final String EXPECTED_EVENT_NAME = "event";
    @RequestMapping("/lists")
    public Object list(){
        try {
            Util.cleanUp();
            UserContext adminUser = new UserContext();
            adminUser.setName(Config.ADMIN);
            adminUser.setAffiliation(Config.ORG2);
            adminUser.setMspId(Config.ORG2_MSP);
           // File f = new File (this.getClass().getResource("/ca.crt").getPath());
            String certficate = new String (IOUtils.toByteArray(this.getClass().getResourceAsStream("/ca.crt")),"UTF-8");
            Properties properties = new Properties();
            properties.put("pemBytes", certficate.getBytes());
           // properties.setProperty("pemFile", f.getAbsolutePath());
            properties.setProperty("allowAllHostNames", "true");
            CAClient caclient=new  CAClient(Config.CA_ORG2_URL, properties);
            caclient.setAdminUserContext(adminUser);
            adminUser =  caclient.enrollAdminUserTLS("admin", "adminpw");
            FabricClient fabClient = new FabricClient(adminUser);
            ChannelClient channelClient = fabClient.createChannelClient(Config.CHANNEL_NAME);
            Channel channel = channelClient.getChannel();
            //Properties propertiess = new Properties();
            /*propertiess.put("pemBytes", Files.readAllBytes(Paths.get("C:\\Users\\xxrib\\Desktop\\linux-fabric\\blockchain-application-using-fabric-java-sdk\\java\\src\\main\\resources\\ca.crt")));
            propertiess.setProperty("sslProvider", "openSSL");
            propertiess.setProperty("negotiationType", "TLS");
            propertiess.setProperty("trustServerCertificate", "true");
            propertiess.setProperty("hostnameOverride", Config.ORG1_PEER_0);*/
            /*File fs = new File ("C:\\Users\\xxrib\\Desktop\\linux-fabric\\blockchain-application-using-fabric-java-sdk\\java\\src\\main\\resources\\peer0.org1.example.com.crt");
            String ficate = new String (IOUtils.toByteArray(new FileInputStream(f)),"UTF-8");
            Properties propert = new Properties();
            propert.put("pemBytes", ficate.getBytes());
            propert.setProperty("pemFile", fs.getAbsolutePath());
            propert.setProperty("allowAllHostNames", "true");*/
            //File fp = new File (this.getClass().getResource("/server-org2p0.crt").getPath());
            String certficatep = new String (IOUtils.toByteArray(this.getClass().getResourceAsStream("/server-org2p0.crt")),"UTF-8");
            Properties peer_properties = new Properties();
            peer_properties.put("pemBytes", certficatep.getBytes());
            peer_properties.setProperty("sslProvider", "openSSL");
            peer_properties.setProperty("negotiationType", "TLS");
            Peer peer = fabClient.getInstance().newPeer(Config.ORG2_PEER_0, Config.ORG2_PEER_0_URL,peer_properties);

            // EventHub eventHub = fabClient.getInstance().newEventHub("eventhub01",Config.grpc+ Config.baseUrl+":7053");
            Orderer orderer = fabClient.getInstance().newOrderer(Config.ORDERER_NAME, Config.ORDERER_URL);
            channel.addOrderer(orderer);
            channel.addPeer(peer);
            channel.initialize();
            Logger.getLogger(QueryChaincode.class.getName()).log(Level.INFO, "Querying   ...");
            Collection<ProposalResponse> responsesQuery = channelClient.queryByChainCode("com/src/chaincode/fabcar", "queryCar", new String[] {"CAR9"});
            for (ProposalResponse pres : responsesQuery) {
                String stringResponse = new String(pres.getChaincodeActionResponsePayload());
                return  stringResponse;
            }
            return null;
        } catch (Exception e) {
            e.printStackTrace();
        }
        return null;
    }
    @RequestMapping("/createCar")
    public Object createCar(String key,String make,String model,String colour,String owner){
        try {
            Util.cleanUp();
            String caUrl = Config.CA_ORG1_URL;
            CAClient caClient = new CAClient(caUrl, null);
            // Enroll Admin to Org1MSP
            UserContext adminUserContext = new UserContext();
            adminUserContext.setName(Config.ADMIN);
            adminUserContext.setAffiliation(Config.ORG2);
            adminUserContext.setMspId(Config.ORG2_MSP);
            caClient.setAdminUserContext(adminUserContext);
            String certficate = new String (IOUtils.toByteArray(this.getClass().getResourceAsStream("/ca.crt")),"UTF-8");
            Properties properties = new Properties();
            properties.put("pemBytes", certficate.getBytes());
            properties.setProperty("allowAllHostNames", "true");
            CAClient caclient=new  CAClient(Config.CA_ORG2_URL, properties);
            caclient.setAdminUserContext(adminUserContext);
            adminUserContext =  caclient.enrollAdminUserTLS("admin", "adminpw");
            FabricClient fabClient = new FabricClient(adminUserContext);
            ChannelClient channelClient = fabClient.createChannelClient(Config.CHANNEL_NAME);
            Channel channel = channelClient.getChannel();
            String certficatep = new String (IOUtils.toByteArray(this.getClass().getResourceAsStream("/server-org2p0.crt")),"UTF-8");
            Properties peer_properties = new Properties();
            peer_properties.put("pemBytes", certficatep.getBytes());
            peer_properties.setProperty("sslProvider", "openSSL");
            peer_properties.setProperty("negotiationType", "TLS");
            Peer peer = fabClient.getInstance().newPeer(Config.ORG2_PEER_0, Config.ORG2_PEER_0_URL,peer_properties);
            //EventHub eventHub = fabClient.getInstance().newEventHub("eventhub01", "grpc://localhost:7053");
            String certficatepOder = new String (IOUtils.toByteArray(this.getClass().getResourceAsStream("/order.crt")),"UTF-8");
            Properties peer_propertiesOder = new Properties();
            peer_propertiesOder.put("pemBytes", certficatepOder.getBytes());
            peer_propertiesOder.setProperty("sslProvider", "openSSL");
            peer_propertiesOder.setProperty("negotiationType", "TLS");
            Orderer orderer = fabClient.getInstance().newOrderer(Config.ORDERER_NAME, Config.ORDERER_URL,peer_propertiesOder);

            String certficatep2 = new String (IOUtils.toByteArray(this.getClass().getResourceAsStream("/server-org1p0.crt")),"UTF-8");
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
            ChaincodeID ccid = ChaincodeID.newBuilder().setName(Config.CHAINCODE_1_NAME).build();
            request.setChaincodeID(ccid);
            request.setFcn("createCar");
            String[] arguments = { key, make, model, colour, owner };
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
                ChaincodeResponse.Status status = res.getStatus();
                if (status.getStatus()==200){
                    return "操作成功";
                }
                Logger.getLogger(InvokeChaincode.class.getName()).log(Level.INFO,"Invoked createCar on "+Config.CHAINCODE_1_NAME + ". Status - " + status);
            }

        } catch (Exception e) {
            e.printStackTrace();
        }
        return "操作失败";
    }
    @RequestMapping("/changOwner")
    public Object changOwner(String key,String newOwner){
        try {
            Util.cleanUp();
            String caUrl = Config.CA_ORG1_URL;
            CAClient caClient = new CAClient(caUrl, null);
            // Enroll Admin to Org1MSP
            UserContext adminUserContext = new UserContext();
            adminUserContext.setName(Config.ADMIN);
            adminUserContext.setAffiliation(Config.ORG2);
            adminUserContext.setMspId(Config.ORG2_MSP);
            caClient.setAdminUserContext(adminUserContext);
            String certficate = new String (IOUtils.toByteArray(this.getClass().getResourceAsStream("/ca.crt")),"UTF-8");
            Properties properties = new Properties();
            properties.put("pemBytes", certficate.getBytes());
           // properties.setProperty("pemFile", f.getAbsolutePath());
            properties.setProperty("allowAllHostNames", "true");
            CAClient caclient=new  CAClient(Config.CA_ORG2_URL, properties);
            caclient.setAdminUserContext(adminUserContext);
            adminUserContext =  caclient.enrollAdminUserTLS("admin", "adminpw");
            FabricClient fabClient = new FabricClient(adminUserContext);
            ChannelClient channelClient = fabClient.createChannelClient(Config.CHANNEL_NAME);
            Channel channel = channelClient.getChannel();
            String certficatep = new String (IOUtils.toByteArray(this.getClass().getResourceAsStream("/server-org2p0.crt")),"UTF-8");
            Properties peer_properties = new Properties();
            peer_properties.put("pemBytes", certficatep.getBytes());
            peer_properties.setProperty("sslProvider", "openSSL");
            peer_properties.setProperty("negotiationType", "TLS");
            Peer peer = fabClient.getInstance().newPeer(Config.ORG2_PEER_0, Config.ORG2_PEER_0_URL,peer_properties);
            //EventHub eventHub = fabClient.getInstance().newEventHub("eventhub01", "grpc://localhost:7053");
            String certficatepOder =new String (IOUtils.toByteArray(this.getClass().getResourceAsStream("/order.crt")),"UTF-8");
            Properties peer_propertiesOder = new Properties();
            peer_propertiesOder.put("pemBytes", certficatepOder.getBytes());
            peer_propertiesOder.setProperty("sslProvider", "openSSL");
            peer_propertiesOder.setProperty("negotiationType", "TLS");
            Orderer orderer = fabClient.getInstance().newOrderer(Config.ORDERER_NAME, Config.ORDERER_URL,peer_propertiesOder);

            String certficatep2 = new String (IOUtils.toByteArray(this.getClass().getResourceAsStream("/server-org1p0.crt")),"UTF-8");
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
            ChaincodeID ccid = ChaincodeID.newBuilder().setName(Config.CHAINCODE_1_NAME).build();
            request.setChaincodeID(ccid);
            request.setFcn("changeCarOwner");
            String[] arguments = { key, newOwner };
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
                ChaincodeResponse.Status status = res.getStatus();
                if (status.getStatus()==200){
                    return "操作成功";
                }
                Logger.getLogger(InvokeChaincode.class.getName()).log(Level.INFO,"Invoked createCar on "+Config.CHAINCODE_1_NAME + ". Status - " + status);
            }

        } catch (Exception e) {
            e.printStackTrace();
        }
        return "操作失败";
    }
    @RequestMapping("/deployCC")
    public Object deployCC(){
        try {
            Util.cleanUp();
            CryptoSuite cryptoSuite = CryptoSuite.Factory.getCryptoSuite();

            UserContext org1Admin = new UserContext();
            Enrollment enrollOrg1Admin = Util.getEnrollment("/usr/local/src/fabric-samples/test-network/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp/keystore/priv_sk", null,
                    "/usr/local/src/fabric-samples/test-network/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp/signcerts/Admin@org1.example.com-cert.pem", null);
            org1Admin.setEnrollment(enrollOrg1Admin);
            org1Admin.setMspId(Config.ORG1_MSP);
            org1Admin.setName(Config.ADMIN);

            UserContext org2Admin = new UserContext();
            Enrollment enrollOrg2Admin = Util.getEnrollment("/usr/local/src/fabric-samples/test-network/organizations/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp/keystore/priv_sk", null,
                    "/usr/local/src/fabric-samples/test-network/organizations/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp/signcerts/Admin@org2.example.com-cert.pem", null);
            org2Admin.setEnrollment(enrollOrg2Admin);
            org2Admin.setMspId(Config.ORG2_MSP);
            org2Admin.setName(Config.ADMIN);
            FabricClient fabClient = new FabricClient(org1Admin);
            HFClient client = fabClient.getInstance();
            Channel mychannel = fabClient.getInstance().newChannel("mychannel");
            Properties orderer1Prop = new Properties();
            orderer1Prop.setProperty("pemFile", "/usr/local/src/fabric-samples/test-network/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem");
            orderer1Prop.setProperty("sslProvider", "openSSL");
            orderer1Prop.setProperty("negotiationType", "TLS");
            orderer1Prop.setProperty("hostnameOverride", "orderer.example.com");
            orderer1Prop.setProperty("trustServerCertificate", "true");
            orderer1Prop.put("grpc.NettyChannelBuilderOption.maxInboundMessageSize", 9000000);
            Orderer orderer = client.newOrderer("orderer.example.com", "grpcs://orderer.example.com:7050", orderer1Prop);

            Properties peer1Prop = new Properties();
            peer1Prop.setProperty("pemFile", "/usr/local/src/fabric-samples/test-network/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp/tlscacerts/tlsca.org1.example.com-cert.pem");
            peer1Prop.setProperty("sslProvider", "openSSL");
            peer1Prop.setProperty("negotiationType", "TLS");
            peer1Prop.setProperty("hostnameOverride", "peer0.org1.example.com");
            peer1Prop.setProperty("trustServerCertificate", "true");
            peer1Prop.put("grpc.NettyChannelBuilderOption.maxInboundMessageSize", 9000000);
            Peer peer = client.newPeer("peer0.org1.example.com", "grpcs://peer0.org1.example.com:7051", peer1Prop);

            Properties peer2Prop = new Properties();
            peer2Prop.setProperty("pemFile", "/usr/local/src/fabric-samples/test-network/organizations/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp/tlscacerts/tlsca.org2.example.com-cert.pem");
            peer2Prop.setProperty("sslProvider", "openSSL");
            peer2Prop.setProperty("negotiationType", "TLS");
            peer2Prop.setProperty("hostnameOverride", "peer0.org2.example.com");
            peer2Prop.setProperty("trustServerCertificate", "true");
            peer2Prop.put("grpc.NettyChannelBuilderOption.maxInboundMessageSize", 9000000);
            Peer peer2 = client.newPeer("peer0.org2.example.com", "grpcs://peer0.org2.example.com:9051", peer2Prop);
            mychannel.addOrderer(orderer);
            mychannel.addPeer(peer2);
            mychannel.addPeer(peer);
            mychannel.initialize();

            List<Peer> org1Peers = new ArrayList<Peer>();
            org1Peers.add(peer);

            List<Peer> org2Peers = new ArrayList<Peer>();
            org2Peers.add(peer2);

            String package1 = fabClient.deployChainCode(Config.CHAINCODE_1_NAME,
                    Config.CHAINCODE_1_PATH, Config.CHAINCODE_ROOT_DIR, TransactionRequest.Type.GO_LANG,
                    Config.CHAINCODE_1_VERSION, org1Peers);

            ChannelClient channelClient = new ChannelClient(mychannel.getName(), mychannel, fabClient);

            String[] arguments = { "" };
            CompletableFuture<BlockEvent.TransactionEvent> init = channelClient.instantiateChainCode(Config.CHAINCODE_1_NAME, Config.CHAINCODE_1_VERSION,
                    Config.CHAINCODE_1_PATH, TransactionRequest.Type.GO_LANG.toString(), "init", arguments, null, package1, org1Peers);

            channelClient.verifyByCheckCommitReadinessStatus(Config.CHAINCODE_1_NAME,true,org1Peers,new HashSet<>(Arrays.asList(Config.ORG1_MSP)), // Approved
                    new HashSet<>(Arrays.asList(Config.ORG2_MSP)));

            fabClient.getInstance().setUserContext(org2Admin);
            channelClient=new ChannelClient(mychannel.getName(), mychannel, fabClient);

            String package2 = fabClient.deployChainCode(Config.CHAINCODE_1_NAME,
                    Config.CHAINCODE_1_PATH, Config.CHAINCODE_ROOT_DIR, TransactionRequest.Type.GO_LANG,
                    Config.CHAINCODE_1_VERSION, org2Peers);

            CompletableFuture<BlockEvent.TransactionEvent> inits = channelClient.instantiateChainCode(Config.CHAINCODE_1_NAME, Config.CHAINCODE_1_VERSION,
                    Config.CHAINCODE_1_PATH, TransactionRequest.Type.GO_LANG.toString(), "init", arguments, null, package2, org2Peers);
            channelClient.verifyByCheckCommitReadinessStatus(Config.CHAINCODE_1_NAME,true,org2Peers,new HashSet<>(Arrays.asList(Config.ORG1_MSP,Config.ORG2_MSP)), // Approved
                    Collections.emptySet());
            channelClient.verifyByCheckCommitReadinessStatus(Config.CHAINCODE_1_NAME,true,org1Peers,new HashSet<>(Arrays.asList(Config.ORG1_MSP,Config.ORG2_MSP)), // Approved
                    Collections.emptySet());
            //沉睡很关键  由于都是回调  可能审批未完成
            Thread.sleep(10000);
            Logger.getLogger(InvokeChaincode.class.getName()).log(Level.INFO,org1Peers.toString());
            CompletableFuture<BlockEvent.TransactionEvent> future2 = channelClient.commitChaincodeDefinitionRequest(Config.CHAINCODE_1_NAME, true,Arrays.asList(peer,peer2));

            //channelClient2.chainCodeInit("init",true,Config.CHAINCODE_1_NAME, TransactionRequest.Type.GO_LANG);
            return "success";
        } catch (Exception e) {
            e.printStackTrace();
        }
        return "操作失败";
    }
    @RequestMapping("initCC")
    public Object init(String codeName) {
        try {
            UserContext org1Admin = new UserContext();
            Enrollment enrollOrg1Admin = Util.getEnrollment("/usr/local/src/fabric-samples/test-network/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp/keystore/priv_sk", null,
                    "/usr/local/src/fabric-samples/test-network/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp/signcerts/Admin@org1.example.com-cert.pem", null);
            org1Admin.setEnrollment(enrollOrg1Admin);
            org1Admin.setMspId(Config.ORG1_MSP);
            org1Admin.setName(Config.ADMIN);

            UserContext org2Admin = new UserContext();
            Enrollment enrollOrg2Admin = Util.getEnrollment("/usr/local/src/fabric-samples/test-network/organizations/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp/keystore/priv_sk", null,
                    "/usr/local/src/fabric-samples/test-network/organizations/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp/signcerts/Admin@org2.example.com-cert.pem", null);
            org2Admin.setEnrollment(enrollOrg2Admin);
            org2Admin.setMspId(Config.ORG2_MSP);
            org2Admin.setName(Config.ADMIN);
            FabricClient fabClient = new FabricClient(org1Admin);
            HFClient client = fabClient.getInstance();
            Channel mychannel = fabClient.getInstance().newChannel("mychannel");
            Properties orderer1Prop = new Properties();
            orderer1Prop.setProperty("pemFile", "/usr/local/src/fabric-samples/test-network/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem");
            orderer1Prop.setProperty("sslProvider", "openSSL");
            orderer1Prop.setProperty("negotiationType", "TLS");
            orderer1Prop.setProperty("hostnameOverride", "orderer.example.com");
            orderer1Prop.setProperty("trustServerCertificate", "true");
            orderer1Prop.put("grpc.NettyChannelBuilderOption.maxInboundMessageSize", 9000000);
            Orderer orderer = client.newOrderer("orderer.example.com", "grpcs://orderer.example.com:7050", orderer1Prop);

            Properties peer1Prop = new Properties();
            peer1Prop.setProperty("pemFile", "/usr/local/src/fabric-samples/test-network/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp/tlscacerts/tlsca.org1.example.com-cert.pem");
            peer1Prop.setProperty("sslProvider", "openSSL");
            peer1Prop.setProperty("negotiationType", "TLS");
            peer1Prop.setProperty("hostnameOverride", "peer0.org1.example.com");
            peer1Prop.setProperty("trustServerCertificate", "true");
            peer1Prop.put("grpc.NettyChannelBuilderOption.maxInboundMessageSize", 9000000);
            Peer peer = client.newPeer("peer0.org1.example.com", "grpcs://peer0.org1.example.com:7051", peer1Prop);

            Properties peer2Prop = new Properties();
            peer2Prop.setProperty("pemFile", "/usr/local/src/fabric-samples/test-network/organizations/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp/tlscacerts/tlsca.org2.example.com-cert.pem");
            peer2Prop.setProperty("sslProvider", "openSSL");
            peer2Prop.setProperty("negotiationType", "TLS");
            peer2Prop.setProperty("hostnameOverride", "peer0.org2.example.com");
            peer2Prop.setProperty("trustServerCertificate", "true");
            peer2Prop.put("grpc.NettyChannelBuilderOption.maxInboundMessageSize", 9000000);
            Peer peer2 = client.newPeer("peer0.org2.example.com", "grpcs://peer0.org2.example.com:9051", peer2Prop);
            mychannel.addOrderer(orderer);
            mychannel.addPeer(peer2);
            mychannel.addPeer(peer);
            mychannel.initialize();

            //init
            ChannelClient channelClient = new ChannelClient(mychannel.getName(), mychannel, fabClient);
            channelClient.chainCodeInit("initLedger",org1Admin,true,codeName, TransactionRequest.Type.GO_LANG);
        } catch (Exception e) {
            return "操作失败";
        }
        return "操作成功";
    }
}

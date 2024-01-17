package org.example.util;

import com.google.protobuf.ByteString;
import org.apache.commons.compress.utils.IOUtils;
import org.example.chaincode.invocation.InvokeChaincode;
import org.example.chaincode.invocation.QueryCar;
import org.example.client.CAClient;
import org.example.client.ChannelClient;
import org.example.client.FabricClient;
import org.example.config.Config;
import org.example.model.CodeInfo;
import org.example.user.UserContext;
import org.hyperledger.fabric.protos.common.Configtx;
import org.hyperledger.fabric.protos.peer.Query;
import org.hyperledger.fabric.sdk.*;
import org.hyperledger.fabric.sdk.exception.CryptoException;
import org.hyperledger.fabric.sdk.exception.InvalidArgumentException;
import org.hyperledger.fabric.sdk.exception.ProposalException;
import org.hyperledger.fabric_ca.sdk.HFCAIdentity;
import org.hyperledger.fabric_ca.sdk.exception.IdentityException;

import java.io.File;
import java.io.FileInputStream;
import java.io.IOException;
import java.lang.reflect.InvocationTargetException;
import java.util.*;

public class FabricUtils {
    public static Collection<Peer> getPeers(String channelName)  {
        ChannelClient client = getChannelClient(channelName, "invoke");
        client.getChannel().getDiscoveredChaincodeNames();
        return client.getChannel().getPeers();
    }
    public static Integer getChannelCodes(String channelName) {
        try {
            HFClient client = getAdmin();
            File fp = new File (QueryCar.class.getResource("/ca.crt").getPath());
            String certficatep = new String (IOUtils.toByteArray(new FileInputStream(fp)),"UTF-8");
            Properties peer_properties = new Properties();
            peer_properties.put("pemBytes", certficatep.getBytes());
            peer_properties.setProperty("sslProvider", "openSSL");
            peer_properties.setProperty("negotiationType", "TLS");
            Peer peer =client.newPeer(Config.ORG2_PEER_0, Config.ORG2_PEER_0_URL,peer_properties);
            Collection<LifecycleQueryInstalledChaincodesProposalResponse> resp = client.sendLifecycleQueryInstalledChaincodes(client.newLifecycleQueryInstalledChaincodesRequest(), Arrays.asList(peer));
            Integer num=0;
            for (LifecycleQueryInstalledChaincodesProposalResponse x: resp) {
                if (x.getStatus().getStatus()==200) {
                    num+=1;
                }
            }
            return num;
        } catch (Exception e) {
            e.printStackTrace();
        }
        return null;
    }
    public static Long getHeight(String channelName) throws InvalidArgumentException, ProposalException {
        ChannelClient client = getChannelClient(channelName, "query");
        BlockchainInfo info = client.getChannel().queryBlockchainInfo();
        return info.getHeight();
    }
    public static BlockInfo getBlockInfo(String channelName,Integer num) throws InvalidArgumentException, ProposalException {
        ChannelClient client = getChannelClient(channelName, "query");
        BlockInfo info = client.getChannel().queryBlockByNumber(num);
        return info;
    }
    public static Set<String> getChannels(){
        try {
            HFClient client = getPeer();
            File fp = new File (QueryCar.class.getResource("/ca.crt").getPath());
            String certficatep = new String (IOUtils.toByteArray(new FileInputStream(fp)),"UTF-8");
            Properties peer_properties = new Properties();
            peer_properties.put("pemBytes", certficatep.getBytes());
            peer_properties.setProperty("sslProvider", "openSSL");
            peer_properties.setProperty("negotiationType", "TLS");
            Peer peer =client.newPeer(Config.ORG2_PEER_0, Config.ORG2_PEER_0_URL,peer_properties);
            Set<String> set = client.queryChannels(peer);
            return set;
        } catch (Exception e){
            e.printStackTrace();
        }
        return null;
    }
    public static Collection<HFCAIdentity> getUsers(){
        Util.cleanUp();
        try {
            /*Org1MSP*/
           /* UserContext adminUser1 = new UserContext();
            adminUser1.setName(Config.ADMIN);
            adminUser1.setAffiliation(Config.ORG2);
            adminUser1.setMspId(Config.ORG2_MSP);
            File f1 = new File (QueryCar.class.getResource("/ca1.crt").getPath());
            String certficate1 = new String (IOUtils.toByteArray(new FileInputStream(f1)),"UTF-8");
            Properties properties1 = new Properties();
            properties1.put("pemBytes", certficate1.getBytes());
            properties1.setProperty("pemFile", f1.getAbsolutePath());
            properties1.setProperty("allowAllHostNames", "true");
            CAClient caclient1=new  CAClient(Config.CA_ORG2_URL, properties1);
            caclient1.setAdminUserContext(adminUser1);
            Collection<HFCAIdentity> identities1 = caclient1.getInstance().getHFCAIdentities(adminUser1);*/
            /*Org2MSP*/
            UserContext adminUser = new UserContext();
            adminUser.setName(Config.ADMIN);
            adminUser.setAffiliation(Config.ORG2);
            adminUser.setMspId(Config.ORG2_MSP);
            File f = new File (QueryCar.class.getResource("/ca.crt").getPath());
            String certficate = new String (IOUtils.toByteArray(new FileInputStream(f)),"UTF-8");
            Properties properties = new Properties();
            properties.put("pemBytes", certficate.getBytes());
            properties.setProperty("pemFile", f.getAbsolutePath());
            properties.setProperty("allowAllHostNames", "true");
            CAClient caclient=new  CAClient(Config.CA_ORG2_URL, properties);
            caclient.setAdminUserContext(adminUser);
            adminUser =  caclient.enrollAdminUserTLS("admin", "adminpw");
            Collection<HFCAIdentity> identities = caclient.getInstance().getHFCAIdentities(adminUser);
            return identities;
        } catch (Exception e) {
            e.printStackTrace();
        }
        return null;
    }
    private static HFClient getPeer(){
        Util.cleanUp();
        try {
            UserContext adminUser = new UserContext();
            adminUser.setName(Config.ADMIN);
            adminUser.setAffiliation(Config.ORG2);
            adminUser.setMspId(Config.ORG2_MSP);
            File f = new File (QueryCar.class.getResource("/ca.crt").getPath());
            String certficate = new String (IOUtils.toByteArray(new FileInputStream(f)),"UTF-8");
            Properties properties = new Properties();
            properties.put("pemBytes", certficate.getBytes());
            properties.setProperty("pemFile", f.getAbsolutePath());
            properties.setProperty("allowAllHostNames", "true");
            CAClient caclient=new  CAClient(Config.CA_ORG2_URL, properties);
            caclient.setAdminUserContext(adminUser);
            adminUser =  caclient.enrollAdminUserTLS("admin", "adminpw");
            FabricClient fabClient = new FabricClient(adminUser);

            return fabClient.getInstance();
        } catch (Exception e) {
            e.printStackTrace();
        }
        return  null;
    }

    private static HFClient getAdmin(){
        Util.cleanUp();
        try {
            UserContext org2Admin = new UserContext();
            Enrollment enrollOrg2Admin = Util.getEnrollment("C:\\Users\\13202\\Desktop\\linux-fabric\\blockchain-application-using-fabric-java-sdk\\java\\src\\main\\resources\\key2\\org2_key", null,
                    "C:\\Users\\13202\\Desktop\\linux-fabric\\blockchain-application-using-fabric-java-sdk\\java\\src\\main\\resources\\admin-cert\\admin2.crt", null);
            org2Admin.setEnrollment(enrollOrg2Admin);
            org2Admin.setMspId(Config.ORG2_MSP);
            org2Admin.setName(Config.ADMIN);
            FabricClient fabClient = new FabricClient(org2Admin);
            return fabClient.getInstance();
        } catch (Exception e) {
            e.printStackTrace();
        }
        return  null;
    }
    private static ChannelClient getChannelClient(String channelName,String opreate){
        Util.cleanUp();
        try {
            UserContext adminUser = new UserContext();
            adminUser.setName(Config.ADMIN);
            adminUser.setAffiliation(Config.ORG2);
            adminUser.setMspId(Config.ORG2_MSP);
            File f = new File (QueryCar.class.getResource("/ca.crt").getPath());
            String certficate = new String (IOUtils.toByteArray(new FileInputStream(f)),"UTF-8");
            Properties properties = new Properties();
            properties.put("pemBytes", certficate.getBytes());
            properties.setProperty("pemFile", f.getAbsolutePath());
            properties.setProperty("allowAllHostNames", "true");
            CAClient caclient=new  CAClient(Config.CA_ORG2_URL, properties);
            caclient.setAdminUserContext(adminUser);
            adminUser =  caclient.enrollAdminUserTLS("admin", "adminpw");
            FabricClient fabClient = new FabricClient(adminUser);
            ChannelClient channelClient = fabClient.createChannelClient(channelName);
            Channel channel = channelClient.getChannel();

            File fp = new File (QueryCar.class.getResource("/ca.crt").getPath());
            String certficatep = new String (IOUtils.toByteArray(new FileInputStream(fp)),"UTF-8");
            Properties peer_properties = new Properties();
            peer_properties.put("pemBytes", certficatep.getBytes());
            peer_properties.setProperty("sslProvider", "openSSL");
            peer_properties.setProperty("negotiationType", "TLS");
            Peer peer = fabClient.getInstance().newPeer(Config.ORG2_PEER_0, Config.ORG2_PEER_0_URL,peer_properties);
            if (opreate.equals("invoke")){
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
                channel.addOrderer(orderer);
                channel.addPeer(peer2);
            }
            channel.addPeer(peer);
            channel.initialize();
            return channelClient;
        } catch (Exception e) {
            e.printStackTrace();
        }
        return null;
    }
}

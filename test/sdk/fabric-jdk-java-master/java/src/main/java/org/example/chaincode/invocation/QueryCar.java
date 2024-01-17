package org.example.chaincode.invocation;

import org.apache.commons.compress.utils.IOUtils;
import org.bouncycastle.openssl.PEMWriter;
import org.example.client.CAClient;
import org.example.client.ChannelClient;
import org.example.client.FabricClient;
import org.example.config.Config;
import org.example.user.UserContext;
import org.example.util.Util;
import org.hyperledger.fabric.sdk.*;
import org.hyperledger.fabric_ca.sdk.EnrollmentRequest;
import org.hyperledger.fabric_ca.sdk.HFCAClient;
import org.hyperledger.fabric_ca.sdk.exception.EnrollmentException;
import org.hyperledger.fabric_ca.sdk.exception.InvalidArgumentException;

import javax.net.ssl.*;
import java.io.File;
import java.io.FileInputStream;
import java.io.IOException;
import java.io.StringWriter;
import java.security.KeyManagementException;
import java.security.NoSuchAlgorithmException;
import java.security.PrivateKey;
import java.security.cert.X509Certificate;
import java.util.Collection;
import java.util.Properties;
import java.util.logging.Level;
import java.util.logging.Logger;

import static java.nio.charset.StandardCharsets.UTF_8;

public class QueryCar {
    private static final byte[] EXPECTED_EVENT_DATA = "!".getBytes(UTF_8);
    private static final String EXPECTED_EVENT_NAME = "event";


    static String getPEMStringFromPrivateKey(PrivateKey privateKey) throws IOException {
        StringWriter pemStrWriter = new StringWriter();
        PEMWriter pemWriter = new PEMWriter(pemStrWriter);

        pemWriter.writeObject(privateKey);

        pemWriter.close();

        return pemStrWriter.toString();
    }


    public static HFCAClient enableTLS(HFCAClient hfcaclient , String uesrName , String userpasswd , String orgName , String caHsotIp) {
        final EnrollmentRequest enrollmentRequestTLS  = new EnrollmentRequest();
        enrollmentRequestTLS.addHost(caHsotIp);
        enrollmentRequestTLS.setProfile("tls");
        try {
            final Enrollment enroll =hfcaclient.enroll(uesrName, userpasswd,enrollmentRequestTLS);
            final String tlsCertPEM = enroll.getCert();
            final String tlsKeyPEM = getPEMStringFromPrivateKey(enroll.getKey());
            final Properties tlsProperties = new Properties ();
            tlsProperties.put("clientKeyBytes", tlsKeyPEM.getBytes(UTF_8));
            tlsProperties.put("clientCertBytes", tlsCertPEM.getBytes(UTF_8));

        } catch (EnrollmentException | InvalidArgumentException e) {
            // TODO Auto-generated catch block
            e.printStackTrace();
        } catch (IOException e) {
            // TODO Auto-generated catch block
            e.printStackTrace();
        }
        return hfcaclient;

    }

    static {
        disableSslVerification();
    }

    private static void disableSslVerification() {
        try
        {
            // Create a trust manager that does not validate certificate chains
            TrustManager[] trustAllCerts = new TrustManager[] {new X509TrustManager() {
                public java.security.cert.X509Certificate[] getAcceptedIssuers() {
                    return null;
                }
                public void checkClientTrusted(X509Certificate[] certs, String authType) {
                }
                public void checkServerTrusted(X509Certificate[] certs, String authType) {
                }
            }
            };

            // Install the all-trusting trust manager
            SSLContext sc = SSLContext.getInstance("SSL");
            sc.init(null, trustAllCerts, new java.security.SecureRandom());
            HttpsURLConnection.setDefaultSSLSocketFactory(sc.getSocketFactory());

            // Create all-trusting host name verifier
            HostnameVerifier allHostsValid = new HostnameVerifier() {
                public boolean verify(String hostname, SSLSession session) {
                    return true;
                }
            };

            // Install the all-trusting host verifier
            HttpsURLConnection.setDefaultHostnameVerifier(allHostsValid);
        } catch (NoSuchAlgorithmException e) {
            e.printStackTrace();
        } catch (KeyManagementException e) {
            e.printStackTrace();
        }
    }



    public static void main(String args[]) {
        try {
            Util.cleanUp();
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
            ChannelClient channelClient = fabClient.createChannelClient(Config.CHANNEL_NAME);
            Channel channel = channelClient.getChannel();

            File fp = new File (QueryCar.class.getResource("/ca.crt").getPath());
            String certficatep = new String (IOUtils.toByteArray(new FileInputStream(fp)),"UTF-8");
            Properties peer_properties = new Properties();
            peer_properties.put("pemBytes", certficatep.getBytes());
            peer_properties.setProperty("sslProvider", "openSSL");
            peer_properties.setProperty("negotiationType", "TLS");
            Peer peer = fabClient.getInstance().newPeer(Config.ORG2_PEER_0, Config.ORG2_PEER_0_URL,peer_properties);
            channel.addPeer(peer);
            channel.initialize();
            System.out.println("channle size"+fabClient.getInstance().queryChannels(peer));
            Logger.getLogger(QueryChaincode.class.getName()).log(Level.INFO, "Querying   ...");
            Collection<ProposalResponse> responsesQuery = channelClient.queryByChainCode("fabcar", "queryCar", new String[] {"CAR9"});
            for (ProposalResponse pres : responsesQuery) {
                String stringResponse = new String(pres.getChaincodeActionResponsePayload());
                Logger.getLogger(QueryChaincode.class.getName()).log(Level.INFO, stringResponse);
            }

        } catch (Exception e) {
            e.printStackTrace();
            
        }
    }
}

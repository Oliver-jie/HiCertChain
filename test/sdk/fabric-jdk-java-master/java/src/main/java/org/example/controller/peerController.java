package org.example.controller;

import com.xiaoleilu.hutool.json.JSONObject;
import org.example.model.IndexInfo;
import org.example.model.PeerInfo;
import org.example.resp.RespUtils;
import org.example.util.FabricUtils;
import org.hyperledger.fabric.sdk.BlockInfo;
import org.hyperledger.fabric.sdk.Peer;
import org.hyperledger.fabric_ca.sdk.HFCAIdentity;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;
import java.util.*;

@RestController
public class peerController {
    @RequestMapping("/peer/index")
    Object getAll(){
        try {
            List<PeerInfo> list = new ArrayList<>();
            Integer chainCodeNum=0;
            /**获取所有通道*/
            Set<String> set = FabricUtils.getChannels();
            for (String channelName:set) {
                Collection<Peer> peers = FabricUtils.getPeers(channelName);
                peers.forEach(x->{
                    if (!list.contains(x)) {
                        PeerInfo peerInfo = new PeerInfo();
                        peerInfo.setPeerName(x.getName());
                        peerInfo.setUrl(x.getUrl());
                        list.add(peerInfo);
                    }
                });
                /**获取channelCode num*/
                chainCodeNum= FabricUtils.getChannelCodes(channelName);

            }
            /**获取所有通道区块高度总和*/
            long num=0;
            for (String x:set) {
                try {
                    num+= FabricUtils.getHeight(x);
                } catch (Exception e) {
                    e.printStackTrace();
                }
            }
            Collection<HFCAIdentity> users = FabricUtils.getUsers();
            IndexInfo info = new IndexInfo();
            info.setPeerNum(list.size());
            info.setChannelNum(set.size());
            info.setBlockHeight(new Long(num).intValue());
            info.setChainCodeNum(chainCodeNum);
            info.setUserNum(users.size());
            return RespUtils.respSuccess("success",info);
        } catch (Exception e) {
            e.printStackTrace();
        }
        return null;
    }
    @RequestMapping("/peer/getPeerInfos")
    Object getInfos(){
        try {
            List<PeerInfo> list = new ArrayList<>();
            /**获取所有通道*/
            Set<String> set = FabricUtils.getChannels();
            for (String channelName:set) {
                Collection<Peer> peers = FabricUtils.getPeers(channelName);
                peers.forEach(x->{
                    if (!list.contains(x)) {
                        PeerInfo peerInfo = new PeerInfo();
                        peerInfo.setPeerName(x.getName());
                        peerInfo.setUrl(x.getUrl());
                        peerInfo.setStatus("ye");
                        list.add(peerInfo);
                    }
                });
            }
            return RespUtils.respSuccess("success",list);
        } catch (Exception e) {
            e.printStackTrace();
        }
        return null;
    }
    @RequestMapping("/peer/getChannels")
    Object getChannels(){

        try {
            Set<String> set = FabricUtils.getChannels();
            List<Map<String,String>> list=new ArrayList<>();
            set.forEach(x->{
                HashMap<String, String> map = new HashMap<>();
                map.put("name",x);
                map.put("status","ye");
                try {
                    Long height= FabricUtils.getHeight(x);
                    map.put("blockHeight",String.valueOf(height));
                } catch (Exception e) {
                    e.printStackTrace();
                }
                list.add(map);
            });
            return RespUtils.respSuccess("success",list);
        } catch (Exception e) {
            e.printStackTrace();
        }
        return null;
    }
    @RequestMapping("/peer/getChannelBlockHeight")
    Object getHeight(JSONObject params){
        if (Objects.isNull(params)){
            return RespUtils.respFail("500","参数不正确",null);
        }
        try {
            long num=0;
                try {
                    num+= FabricUtils.getHeight(params.get("channelName").toString());
                } catch (Exception e) {
                    e.printStackTrace();
                }
            return RespUtils.respSuccess("success",num);
        } catch (Exception e) {
            e.printStackTrace();
        }
        return  null;
    }
    @RequestMapping("/peer/getBlockInfo")
    Object getBlockInfo(@RequestBody JSONObject params){
        if (Objects.isNull(params)){
            return RespUtils.respFail("500","参数不正确",null);
        }
        try {
            BlockInfo info = FabricUtils.getBlockInfo((String) params.get("channelName"), (Integer) params.get("num"));
            String data = info.getBlock().getData().toString();
            return RespUtils.respSuccess("success",data);

        } catch (Exception e) {
            e.printStackTrace();
        }
        return  null;
    }
    @RequestMapping("/peer/getUserInfo")
    Object getUserInfo(){
        /*if (Objects.isNull(params)){
            return RespUtils.respFail("500","参数不正确",null);
        }*/
        try {
            return RespUtils.respSuccess("success",FabricUtils.getUsers());

        } catch (Exception e) {
            e.printStackTrace();
        }
        return  null;
    }
}

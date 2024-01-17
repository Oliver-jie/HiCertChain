package org.example.resp;

public class RespUtils {
    public static Resp respSuccess(String msg,Object data){
        return new Resp("200",msg,data);
    }
    public static Resp respFail(String code,String msg,Object data){
        return new Resp(code,msg,data);
    }
}
class Resp{
    private String code;
    private String msg;
    private Object data;

    public Resp() {
    }

    public Resp(String code, String msg) {
        this.code = code;
        this.msg = msg;
    }

    public Resp(String msg, Object data) {
        this.msg = msg;
        this.data = data;
    }

    public Resp(String code, String msg, Object data) {
        this.code = code;
        this.msg = msg;
        this.data = data;
    }

    public String getCode() {
        return code;
    }

    public void setCode(String code) {
        this.code = code;
    }

    public String getMsg() {
        return msg;
    }

    public void setMsg(String msg) {
        this.msg = msg;
    }

    public Object getData() {
        return data;
    }

    public void setData(Object data) {
        this.data = data;
    }
}
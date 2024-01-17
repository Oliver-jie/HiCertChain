/*
 * SPDX-License-Identifier: Apache-2.0
 */

package com.xxribd.fabric.samples.Wish;

import java.util.ArrayList;
import java.util.List;

import org.hyperledger.fabric.contract.Context;
import org.hyperledger.fabric.contract.ContractInterface;
import org.hyperledger.fabric.contract.annotation.Contact;
import org.hyperledger.fabric.contract.annotation.Contract;
import org.hyperledger.fabric.contract.annotation.Default;
import org.hyperledger.fabric.contract.annotation.Info;
import org.hyperledger.fabric.contract.annotation.License;
import org.hyperledger.fabric.contract.annotation.Transaction;
import org.hyperledger.fabric.shim.ChaincodeException;
import org.hyperledger.fabric.shim.ChaincodeStub;
import org.hyperledger.fabric.shim.ledger.KeyValue;
import org.hyperledger.fabric.shim.ledger.QueryResultsIterator;

import com.owlike.genson.Genson;

/**
 * Java implementation of the Fabric Car Contract described in the Writing Your
 * First Application tutorial
 */
@Contract(
        name = "FabWish",
        info = @Info(
                title = "FabWish contract",
                description = "The hyperlegendary wish contract",
                version = "0.0.1-SNAPSHOT",
                license = @License(
                        name = "Apache 2.0 License",
                        url = "http://www.apache.org/licenses/LICENSE-2.0.html"),
                contact = @Contact(
                        email = "f.carr@example.com",
                        name = "F Carr",
                        url = "https://hyperledger.example.com")))
@Default
public final class FabWish implements ContractInterface {

    private final Genson genson = new Genson();

    private enum FabWishErrors {
        WISH_NOT_FOUND,
        WISH_ALREADY_EXISTS
    }

    /**
     * Retrieves a car with the specified key from the ledger.
     *
     * @param ctx the transaction context
     * @param key the key
     * @return the Car found on the ledger if there was one
     */
    @Transaction()
    public WishInfo queryCar(final Context ctx, final String key) {
        ChaincodeStub stub = ctx.getStub();
        String wishState = stub.getStringState(key);

        if (wishState.isEmpty()) {
            String errorMessage = String.format("Car %s does not exist", key);
            System.out.println(errorMessage);
            throw new ChaincodeException(errorMessage, FabWishErrors.WISH_NOT_FOUND.toString());
        }

        WishInfo car = genson.deserialize(wishState, WishInfo.class);

        return car;
    }

    /**
     * Creates some initial Cars on the ledger.
     *
     * @param ctx the transaction context
     */
    @Transaction()
    public void initLedger(final Context ctx) {
        ChaincodeStub stub = ctx.getStub();

        String[] carData = {
                /*"{ \"make\": \"Toyota\", \"model\": \"Prius\", \"color\": \"blue\", \"owner\": \"Tomoko\" }",
                "{ \"make\": \"Ford\", \"model\": \"Mustang\", \"color\": \"red\", \"owner\": \"Brad\" }",
                "{ \"make\": \"Hyundai\", \"model\": \"Tucson\", \"color\": \"green\", \"owner\": \"Jin Soo\" }",
                "{ \"make\": \"Volkswagen\", \"model\": \"Passat\", \"color\": \"yellow\", \"owner\": \"Max\" }",
                "{ \"make\": \"Tesla\", \"model\": \"S\", \"color\": \"black\", \"owner\": \"Adrian\" }",
                "{ \"make\": \"Peugeot\", \"model\": \"205\", \"color\": \"purple\", \"owner\": \"Michel\" }",
                "{ \"make\": \"Chery\", \"model\": \"S22L\", \"color\": \"white\", \"owner\": \"Aarav\" }",
                "{ \"make\": \"Fiat\", \"model\": \"Punto\", \"color\": \"violet\", \"owner\": \"Pari\" }",
                "{ \"make\": \"Tata\", \"model\": \"nano\", \"color\": \"indigo\", \"owner\": \"Valeria\" }",
                "{ \"make\": \"Holden\", \"model\": \"Barina\", \"color\": \"brown\", \"owner\": \"Shotaro\" }"*/
        };

        /*for (int i = 0; i < carData.length; i++) {
            String key = String.format("CAR%d", i);

            WishInfo car = genson.deserialize(carData[i], WishInfo.class);
            String carState = genson.serialize(car);
            stub.putStringState(key, carState);
        }*/
    }


    @Transaction()
    public WishInfo createCar(final Context ctx, final String key, final String name, final String owner,
                              final String phone, final String wishId,final String data) {
        ChaincodeStub stub = ctx.getStub();
        /**key  就是wishId*/
        String carState = stub.getStringState(key);
        if (!carState.isEmpty()) {
            String errorMessage = String.format(key);
            System.out.println(errorMessage);
            throw new ChaincodeException(errorMessage, FabWishErrors.WISH_ALREADY_EXISTS.toString());
        }

        WishInfo wish = new WishInfo(name,data, owner, phone, wishId);
        carState = genson.serialize(wish);
        stub.putStringState(key, carState);

        return wish;
    }


    @Transaction()
    public String queryAllCars(final Context ctx) {
        ChaincodeStub stub = ctx.getStub();

        final String startKey = "";
        final String endKey = "";
        List<WishQueryResult> queryResults = new ArrayList<>();

        QueryResultsIterator<KeyValue> results = stub.getStateByRange(startKey, endKey);

        for (KeyValue result: results) {
            WishInfo wish = genson.deserialize(result.getStringValue(), WishInfo.class);
            queryResults.add(new WishQueryResult(result.getKey(), wish));
        }

        final String response = genson.serialize(queryResults);

        return response;
    }

    /**
     * Changes the owner of a car on the ledger.
     *
     * @param ctx the transaction context
     * @param key the key
     * @param newOwner the new owner
     * @return the updated Car
     */
    @Transaction()
    public WishInfo changeCarOwner(final Context ctx, final String key, final String newOwner) {
        ChaincodeStub stub = ctx.getStub();

        String carState = stub.getStringState(key);

        if (carState.isEmpty()) {
            String errorMessage = String.format("Car %s does not exist", key);
            System.out.println(errorMessage);
            throw new ChaincodeException(errorMessage, FabWishErrors.WISH_NOT_FOUND.toString());
        }

        WishInfo wish = genson.deserialize(carState, WishInfo.class);

        WishInfo newWish = new WishInfo(wish.getName(), wish.getData(), wish.getPhone(), wish.getWishId(),newOwner);
        String newWishState = genson.serialize(newWish);
        stub.putStringState(key, newWishState);

        return newWish;
    }
}

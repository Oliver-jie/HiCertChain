/*
 * SPDX-License-Identifier: Apache-2.0
 */

package com.xxribd.fabric.samples.Wish;

import java.util.Objects;

import org.hyperledger.fabric.contract.annotation.DataType;
import org.hyperledger.fabric.contract.annotation.Property;

import com.owlike.genson.annotation.JsonProperty;

@DataType()
public final class WishInfo {

    @Property()
    private final String owner;

    @Property()
    private final String phone;

    @Property()
    private final String name;

    @Property()
    private final String wishId;

    @Property()
    private final String data;

    public String getOwner() {
        return owner;
    }

    public String getPhone() {
        return phone;
    }

    public String getName() {
        return name;
    }

    public String getWishId() {
        return wishId;
    }
    public String getData() {
        return data;
    }

    public WishInfo( @JsonProperty("owner") final String owner,@JsonProperty("data") final String data
            , @JsonProperty("name") final String name, @JsonProperty("phone") final String phone, @JsonProperty("wishId") final String wishId) {
        this.owner = owner;
        this.name=name;
        this.phone=phone;
        this.wishId=wishId;
        this.data=data;
    }

    @Override
    public boolean equals(final Object obj) {
        if (this == obj) {
            return true;
        }

        if ((obj == null) || (getClass() != obj.getClass())) {
            return false;
        }

        WishInfo other = (WishInfo) obj;

        return Objects.deepEquals(new String[] {getName(), getOwner(), getPhone(), getWishId(),getData()},
                new String[] {other.getName(), other.getOwner(), other.getPhone(), other.getWishId(),other.getData()});
    }

    @Override
    public int hashCode() {
        return Objects.hash(getName(), getOwner(), getPhone(), getWishId(),getData());
    }

    @Override
    public String toString() {
        return this.getClass().getSimpleName() + "@" + Integer.toHexString(hashCode()) + " [name=" + name + ", owner="
                + owner + ", phone=" + phone + ", wishId=" + wishId + ", data=" + data+ "]";
    }
}

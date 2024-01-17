package org.example.model;

import lombok.Data;
import lombok.EqualsAndHashCode;

import java.util.Objects;

@Data
public class PeerInfo {
    private String peerName;

    private String url;

    private String status;

    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (o == null || getClass() != o.getClass()) return false;
        PeerInfo peerInfo = (PeerInfo) o;
        return Objects.equals(peerName, peerInfo.peerName) &&
                Objects.equals(url, peerInfo.url);
    }

    @Override
    public int hashCode() {
        return Objects.hash(peerName, url);
    }
}

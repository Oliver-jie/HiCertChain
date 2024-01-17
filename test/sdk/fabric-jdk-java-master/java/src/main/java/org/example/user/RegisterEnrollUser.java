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
package org.example.user;

import org.apache.commons.compress.utils.IOUtils;
import org.example.chaincode.invocation.QueryCar;
import org.example.client.CAClient;
import org.example.config.Config;
import org.example.util.Util;

import java.io.File;
import java.io.FileInputStream;
import java.util.Properties;

/**
 * 
 * @author Balaji Kadambi
 *
 */

public class RegisterEnrollUser {

	public static void main(String args[]) {
		try {
			Util.cleanUp();
			File f = new File (QueryCar.class.getResource("/ca.crt").getPath());
			String certficate = new String (IOUtils.toByteArray(new FileInputStream(f)),"UTF-8");
			Properties properties = new Properties();
			properties.put("pemBytes", certficate.getBytes());
			properties.setProperty("pemFile", f.getAbsolutePath());
			properties.setProperty("allowAllHostNames", "true");
			String caUrl = Config.CA_ORG2_URL;
			CAClient caClient = new CAClient(caUrl, properties);
			// Enroll Admin to Org1MSP
			UserContext adminUserContext = new UserContext();
			adminUserContext.setName(Config.ADMIN);
			adminUserContext.setAffiliation(Config.ORG2);
			adminUserContext.setMspId(Config.ORG2_MSP);
			caClient.setAdminUserContext(adminUserContext);
			adminUserContext = caClient.enrollAdminUserTLS(Config.ADMIN, Config.ADMIN_PASSWORD);

			// Register and Enroll user to Org1MSP
			UserContext userContext = new UserContext();
			String name = "my_user777";
			userContext.setName(name);
			userContext.setAffiliation(Config.ORG2);
			userContext.setMspId(Config.ORG2_MSP);

			String eSecret = caClient.registerUser(name, Config.ORG2);

			userContext = caClient.enrollUser(userContext,eSecret);
			System.out.println(userContext);
		} catch (Exception e) {
			e.printStackTrace();
		}
	}

}

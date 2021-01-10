/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package devmon;

import java.io.FileInputStream;
import java.util.Properties;

/**
 *
 * @author mylavari
 */
public class ParamConfigurationUtil {
    //Repository db jdbc config
    public String repodb_jdbc_driver_classname;
    public String repodb_jdbc_connectiion_url;
    public String repodb_jdbc_user;
    public String repodb_jdbc_password;

    //SMTP config
    public String mail_smtp_host;
    public String mail_smtp_port;
    public String mail_smtp_auth;
    public String mail_smtp_starttls_enable;
    public String mail_smtp_from;
    public String mail_smtp_to;
    public String mail_smtp_cc;
    public String mail_smtp_user;
    public String mail_smtp_password;
    
    public  ParamConfigurationUtil(String file) {
        Properties prop = new Properties();
        try {
            prop.load(new FileInputStream(file));
        } catch (Exception ex) {
            System.out.println("Exception in loading " + file + " file:" + ex);
        }
        System.out.println(prop);
        this.repodb_jdbc_driver_classname=prop.getProperty("repodb.jdbc_driver_classname");
        this.repodb_jdbc_connectiion_url=prop.getProperty("repodb.jdbc_connectiion_url");
        this.repodb_jdbc_user=prop.getProperty("repodb.jdbc_user");
        this.repodb_jdbc_password=prop.getProperty("repodb.jdbc_password");
        this.mail_smtp_host=prop.getProperty("mail_smtp_host");
        this.mail_smtp_port=prop.getProperty("mail_smtp_port");        
        this.mail_smtp_auth=prop.getProperty("mail_smtp_auth");
        this.mail_smtp_starttls_enable=prop.getProperty("mail_smtp_starttls_enable");
        this.mail_smtp_from=prop.getProperty("mail_smtp_from");
        this.mail_smtp_to=prop.getProperty("mail_smtp_to");
        this.mail_smtp_cc=prop.getProperty("mail_smtp_cc");
        this.mail_smtp_user=prop.getProperty("mail_smtp_user");
        this.mail_smtp_password=prop.getProperty("mail_smtp_password");
    }
}

package devmon;

import java.sql.*;

import java.time.LocalDateTime;
import java.util.ArrayList;
import java.util.LinkedList;
import java.util.List;
import java.util.Properties;

import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.ThreadPoolExecutor;

import javax.mail.*;
import javax.mail.internet.InternetAddress;
import javax.mail.internet.MimeMessage;

public class Devmon {
    static boolean keep_running  = true;
    static CronQueue queue = new CronQueue();
    static CronTab ct = new CronTab();
    
    public static ParamConfigurationUtil cfg;


    public static void main(String[] args)  {
        cfg = new ParamConfigurationUtil("App.properties");

        ExecutorService Execurot = Executors.newFixedThreadPool(2);
        ThreadPoolExecutor pool = (ThreadPoolExecutor) Execurot;
        Execurot.submit(new Cron2Queue());
        Execurot.submit(new Queue2Workers());
        
    }
    
    static class SQLWorker implements Runnable {
        Mp_info mpinfo;
        public SQLWorker(Mp_info mpinfo) {
            this.mpinfo=mpinfo;
        }
        @Override
        public void run() {
            System.out.println("SQLWorker:" + mpinfo.cmd);

            try {
                Timestamp start_ts = new Timestamp(System.currentTimeMillis());
                Connection cn = DriverManager.getConnection(mpinfo.url,cfg.repodb_jdbc_user,cfg.repodb_jdbc_password);
                ResultSet rs = cn.createStatement().executeQuery(mpinfo.cmd);
                Timestamp end_ts = new Timestamp(System.currentTimeMillis());
                mpinfo.start_ts = start_ts.toString();
                mpinfo.end_ts = end_ts.toString();
                Connection rcn = DriverManager.getConnection(cfg.repodb_jdbc_connectiion_url,cfg.repodb_jdbc_user,cfg.repodb_jdbc_password);
                PreparedStatement pst = rcn.prepareStatement("insert into job_log(epmg_id,epg_id,cmd_id,cron_id,ep_id,start_ts,end_ts,result) values (?,?,?,?,?,?,?,?)");
                pst.setInt(1, mpinfo.epmg_id);
                pst.setInt(2, mpinfo.epg_id);
                pst.setInt(3, mpinfo.cmd_id);
                pst.setInt(4, mpinfo.cron_id);
                pst.setInt(5, mpinfo.ep_id);
                pst.setString(6,start_ts.toString());
                pst.setString(7,end_ts.toString());
                pst.setString(8,rs2jsonString(rs));  
                pst.execute();
                pst.close();
                rcn.close();
                rs.close();
                cn.close();
            } catch (Exception ex) {
                System.out.println("Exception in Queue2Workers : " + ex);
            }
        }
        private String rs2jsonString(ResultSet rs){
            String jsonString="";
            String ColumnName="";
            int numColumns;
            boolean rows_exist;
            try {
                ResultSetMetaData rsmd = rs.getMetaData();
                numColumns = rsmd.getColumnCount();
                rows_exist = false;
                while(rs.next()) {
                    if (rows_exist) jsonString +=",";
                    rows_exist = true;
                    jsonString += "{";
                    for (int i=1; i <= numColumns; i++) {
                        jsonString += rsmd.getColumnName(i) + ":" + rs.getObject(i);
                        if ( i < numColumns ) jsonString +=",";
                    }
                    jsonString += "}";
                }
            } catch (SQLException ex) {
                System.out.println(ex);
            }
            return "{rs:[" + jsonString + "]}";
        }        
    }
    
    static class Queue2Workers implements Runnable {
        @Override
        public void run() {
            System.out.println("Queue2Workers");
            ExecutorService executor = Executors.newFixedThreadPool(100);
            ThreadPoolExecutor pool = (ThreadPoolExecutor) executor;

            try {
                Connection cn = DriverManager.getConnection(cfg.repodb_jdbc_connectiion_url,cfg.repodb_jdbc_user,cfg.repodb_jdbc_password);
                Mp_info mpinfo = new Mp_info();
                String url,cmd;
                while (keep_running) {
                    if (queue.size() > 0 ) {
                        for (int i=0; i < queue.size();i++) {
                            mpinfo.cron_id=queue.pop();
                            System.out.println( "cron_id: " + mpinfo.cron_id);
                            PreparedStatement pst = cn.prepareStatement("select epmg_id,epg_id,cmd_id,cron_id,ep_id,url,cmd FROM mp_info where cron_id=?");
                            pst.setInt(1, mpinfo.cron_id);
                            ResultSet rs= pst.executeQuery();
                            while(rs.next()) {
                                mpinfo.epmg_id=rs.getInt(1);
                                mpinfo.epg_id=rs.getInt(2);                                
                                mpinfo.cmd_id=rs.getInt(3);
                                mpinfo.cron_id=rs.getInt(4);
                                mpinfo.ep_id=rs.getInt(5);
                                mpinfo.url=rs.getString(6);
                                mpinfo.cmd=rs.getString(7);
                                executor.submit(new SQLWorker(mpinfo));
                            }
                            rs.close();
                            pst.close();
                        }
                    } else 
                        System.out.println("Queue2Workers - empty..."); 
                    Thread.sleep(60000);   
                }
            } catch (Exception ex) {
                System.out.println("Exception in Queue2Workers : " + ex);
            }
        }
    }
    static class Cron2Queue implements Runnable {
        
        @Override
        public void run() {
            int id=0;
            String name="at every 10 mins";
            String expr = "*/10 */4";
            //ct.parse(id, name,expr);
            Connection cn;
            Statement st;
            System.out.println("Cron2Queue");
            try {
                cn = DriverManager.getConnection(cfg.repodb_jdbc_connectiion_url,cfg.repodb_jdbc_user,cfg.repodb_jdbc_password);
                cn.setAutoCommit(false);
                st = cn.createStatement();

                ResultSet rs = st.executeQuery("select cron_id id,expr from cron");
                List<CronTab> l = new ArrayList<>();
                while (rs.next()) {
                    CronTab c = new CronTab();
                    c.parse((short)rs.getInt(1),rs.getString(2));
                    l.add(c);
                }
                rs.close();
                st.close();
                cn.close();

                while (keep_running) {
                    long startTime = System.currentTimeMillis();
                    LocalDateTime now = LocalDateTime.now();
                    System.out.println(now);

                    int[] myNum = {now.getMinute(),now.getHour(),now.getDayOfMonth(),now.getMonthValue(),now.getDayOfWeek().getValue() };
                    for (int i = 0; i < l.size(); i++) {
                        int match = 0;
                        for (int j = 0; j < l.get(i).LSIntegers.size(); j++) {
                            if (l.get(i).LSIntegers.get(j).contains(-1)  || l.get(i).LSIntegers.get(j).contains(myNum[j]) )  match++;
                            if (match == 5) {
                                    System.out.println("Cron2Queue: " + l.get(i).getId());
                                    queue.push(l.get(i).getId());
                            }
                        }
                    }
                    long estimatedSleepTime = System.currentTimeMillis() - startTime;
                    System.out.printf("estimatedSleepTime: %d\n",estimatedSleepTime);
                    Thread.sleep(60000-estimatedSleepTime);
                }
            } catch (Exception ex) {
                System.out.println("Exception in Cron2Queue: " + ex.toString());
            }
        }
    }
    static class Mp_info {
	int epmg_id,epg_id,cmd_id,cron_id,ep_id;
	String url,cmd,start_ts,end_ts,result;
    }

    static class CronQueue {
        LinkedList<Integer> queue = new LinkedList<>();
        public synchronized void push(int value){
           this.queue.push(value);
         }
        public synchronized int pop(){
           return this.queue.pop();
        }
        public int size(){
           return this.queue.size();
        }    
    }    
    static class Email { 
        Properties smtp_prop;
        Session session;
        public Email() {
            smtp_prop = new Properties();
            this.smtp_prop.put("mail_smtp_host", cfg.mail_smtp_host);
            this.smtp_prop.put("mail_smtp_port", cfg.mail_smtp_port);
            this.smtp_prop.put("mail_smtp_auth", cfg.mail_smtp_auth);
            this.smtp_prop.put("mail_smtp_starttls_enable", cfg.mail_smtp_starttls_enable);        
            
            session = Session.getDefaultInstance(this.smtp_prop);            
        }
        public void Send(String subject,String body) {
            try{         // Create a default MimeMessage object.
                MimeMessage message = new MimeMessage(session);
                message.setFrom(new InternetAddress(cfg.mail_smtp_from));
                message.addRecipient(Message.RecipientType.TO, new InternetAddress(cfg.mail_smtp_to));
                message.setSubject(subject);
                message.setText(body);
                Transport.send(message);
            } catch (Exception ex) {
                ex.printStackTrace();
            }
        }
    }
}

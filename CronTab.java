/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package devmon;

import java.util.ArrayList;
import java.util.Arrays;
import java.util.HashSet;
import java.util.List;
import java.util.Set;

/**
 *
 * @author mylavari
 */
public class CronTab {
     static final byte CRON_COLUMNS=5;
     enum CXName {MINUTE,HOUR,DOM,MONTH,DOW};
     static byte[][] CXAllowed = {{0,59},{0,23},{1,31},{1,12},{0,6}};
     private int Id;
     public List<Set<Integer>> LSIntegers;

     
        public int getId() {
            return Id;
        }
        public void setId(int Id) {
            this.Id = Id;
        }        
        public List<Set<Integer>> getLSIntegers() {
            return LSIntegers;
        }
        public void setLSIntegers(List<Set<Integer>> LSIntegers) {
            this.LSIntegers = LSIntegers;
        }
        public void parse(short Id,String cronExpr ) {
            this.setId(Id);
            List<String> cronProp = new ArrayList();
            cronProp.addAll(Arrays.asList(cronExpr.split(" ")));

            System.out.println(cronProp);
            if (cronProp.size() < CRON_COLUMNS) for(int i=cronProp.size();i< CRON_COLUMNS;i++) cronProp.add("*");
            if (cronProp.size() > CRON_COLUMNS) for(int i=CRON_COLUMNS;i<cronProp.size();i++) cronProp.remove(i);
        
            Set<Integer> cs;
            List<Set<Integer>> acs = new ArrayList<>();
            for(int i=0;i<CRON_COLUMNS;i++) {
                cs = parse_sub(i,cronProp.get(i));
                acs.add((Set<Integer>) cs);
            }
            setLSIntegers(acs);
        }
        private static Set<Integer> parse_sub(int cron_pos,String s1)
        {
            int hiphen_pos,devide_pos;
            int c1,c2,c3;
            int min_value, max_value;
            String sx;
            min_value=CXAllowed[cron_pos][0];
            max_value=CXAllowed[cron_pos][1];
            List<String> as1 = new ArrayList<>();
            List<String> as2 ;
            Set<Integer> as3 = new HashSet<>();

            s1=s1.replace("  "," ").replace(" ,",",").replace(", ",",");
            as1.addAll(Arrays.asList(s1.split(",")));
            for(int i=0;i < as1.size();i++)
            {
                as2 = new ArrayList<>();
                sx=as1.get(i);
                hiphen_pos = sx.indexOf("-");
                devide_pos = sx.indexOf("/");

                if (devide_pos > 0 || hiphen_pos > 0) {
                    if (sx.contains("*")) 
                        sx=sx.replace("*",String.valueOf(max_value));
                    if (devide_pos > 0 && hiphen_pos < 0) 
                        sx=String.valueOf(min_value)+ "-" + sx;
                    if (devide_pos < 0 && hiphen_pos > 0 ) 
                        sx= sx + "/1";
                    sx=sx.replace("-",",").replace("/",",");
                    as2.addAll(Arrays.asList(sx.split(",")));
                    if (as2.size() > 2) {
                        c1 = Integer.parseInt(as2.get(0));
                        c2 = Integer.parseInt(as2.get(1));
                        c3 = Integer.parseInt(as2.get(2));
                        for(int j=c1; j<= c2; j+=c3)
                            as3.add(j);
                    }
                } else {
                    if (sx.equals("*")) 
                    as3.add(-1);
                }
            }
            return as3;
        }
}

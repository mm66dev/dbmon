import PySimpleGUI as sg
import pymssql

def get_rs_list(dbname,sql,wd):
    print(sql,wd)
    cur = pymssql.connect(server = "(local)\SQLEXPRESS",user="sa",password="***",database="{}".format(dbname)).cursor()
    cur.execute(sql,tuple(wd))
    col_list=[x[0] for x in cur.description]
    print("description:", cur.description)
    data_list=list(cur.fetchall())
    cur.close()
    return [col_list,data_list]

def table_form(dbname,tabname, ws,wd):
    return_row=[]
    sql=f"select * from {tabname}"
    if len(wd) > 0:
       sql=sql + f" where {ws}"
    td_list=get_rs_list(dbname,sql,tuple(wd))
    print("td_list:",td_list[0])
    delayout=[[sg.Table(values=td_list[1],headings=td_list[0],key='-tdlist-',auto_size_columns=False,vertical_scroll_only=False,justification='left',enable_events=True)],
	[sg.Button('Ok', s=(10, 1)), sg.Button('Cancel', s=(10, 1))]
    ]
    twin =sg.Window('Database Table CRUD tool',delayout,size=(820, 640))
    while True:
       event, data = twin.read()
       print(event, data)
       if event in (sg.WIN_CLOSED,'Cancel','Exit'):
          return_row=[]
          break
       if event in ('Ok'):
          break
       if event in ('-tdlist-'):
          row_in_td_list=data['-tdlist-'][0]
          return_row=td_list[1][row_in_td_list]
          print( td_list[1][row_in_td_list] )
          if len(data['-tdlist-']) > 1:
             return_row=[]
             sg.Popup('Can only select one row', keep_on_top=True)
    twin.close()
    return return_row

def clear_de_form(window,fields):
    for f in fields:    
        window[f["name"]].update("")
def de_form(dbname,tabname):
    #print("tabname: {}".format(tabname))
    frame_layout=[]
    cn = pymssql.connect(server = "(local)\SQLEXPRESS",user="sa",password="***",database=dbname)
    cur = cn.cursor(as_dict=True)
    frame_layout=[]
    fields=[]
    cur.execute("""select C.ORDINAL_POSITION cid,C.COLUMN_NAME name, iif((select 1 
	FROM INFORMATION_SCHEMA.TABLE_CONSTRAINTS T
	JOIN INFORMATION_SCHEMA.CONSTRAINT_COLUMN_USAGE CU 
	ON CU.CONSTRAINT_NAME=T.CONSTRAINT_NAME
	WHERE CU.table_name=C.table_name and CU.COLUMN_NAME=C.COLUMN_NAME) is null,0,1) as pk,
    COLUMNPROPERTY(object_id(C.TABLE_SCHEMA+'.'+C.table_name), C.COLUMN_NAME, 'IsIdentity') id
    FROM INFORMATION_SCHEMA.COLUMNS C WHERE C.table_name=%s""",(tabname,))
    for row in cur.fetchall():
        fields.append(row)
    pkf=[]
    idf=[]
    pkf_exist=False
    for f in fields:
        kf=""
        if f["pk"]==1: 
            pkf.append(f["cid"])
            kf=kf + " (pk)"
            
        if f["id"]==1: 
            idf.append(f["cid"])
            kf=kf + " (id)"
                
        frame_layout.append([sg.Text(f["name"].capitalize() + kf, size=(15, 1)), sg.InputText(key=f["name"])])
    layout = [
             [sg.Frame(tabname, frame_layout, font='Any 12', title_color='blue',key="frame1")],
             [sg.Button("Save"),sg.Button("Search"), sg.Button("Delete"),sg.Button("Clear"),sg.Cancel()]
             ]
    window = sg.Window('{} - Table CRUD form'.format(tabname), layout)
    found_flag=True
    while True:
        event, data = window.read()
        print(event, data)
        if event in (sg.WIN_CLOSED,'Cancel','Exit'):
            break
        if event in ("Search"):
            where_data=[ window[f["name"]].get() for f in fields  if len(window[f["name"]].get())> 0 ]
            where_str=' and '.join("["+f["name"]+"]"+"=%s" for f in fields if len(window[f["name"]].get())> 0)
            row_data=table_form(dbname,tabname,where_str,where_data)
            print("row_data:",row_data)
            if len(row_data) != 0:
               found_flag=True
               i=0
               for f in fields:
                   window[f["name"]].update(row_data[i])
                   i=i+1
        elif event in ("Save"):
            try:
                where_str=' and '.join("["+f["name"]+"]"+"=%s" for f in fields if f["cid"] in pkf)
                where_data=[ window[f["name"]].get() for f in fields if f["cid"] in pkf ]
                if found_flag:
                    set_fields=','.join("["+f["name"]+"]" + "=%s" for f in fields if f["cid"] not in pkf + idf)
                    set_data=[window[f["name"]].get() for f in fields if f["cid"] not in pkf + idf]
                    sql="update [" + tabname + "] set " + set_fields + " where " + where_str
                    print(sql,set_data + where_data)
                    cur.execute(sql,tuple(set_data + where_data))
                    cn.commit()
                    found_flag=False
                else:
                    insert_cols=",".join("["+f["name"]+"]" for f in fields if f["cid"] not in idf)
                    insert_parms=",".join("%s" for f in fields if f["cid"] not in idf)
                    insert_data=[window[f["name"]].get() for f in fields  if f["cid"] not in idf ]
                    sql="insert into [" + tabname + "] (" + insert_cols + ") values (" + insert_parms + ")"
                    print(sql,insert_data)
                    cur.execute(sql,tuple(insert_data))
                    cn.commit()
                    found_flag=False
                clear_de_form(window,fields)    
            except Exception as ex:
                sg.popup_ok("Error: {}".format(ex))
        elif event in ("Delete"):
            where_str=' and '.join("["+f["name"]+"]"+"=%s" for f in fields if f["cid"] in pkf)
            where_data=[ window[f["name"]].get() for f in fields if f["cid"] in pkf ]
            sql="delete from [" + tabname + "] where " + where_str
            cur.execute(sql,tuple(where_data))
            cn.commit()
            clear_de_form(window,fields)    
            found_flag=False
        elif event in ("Clear"):
            found_flag=False
            clear_de_form(window,fields)    
    cur.close()

    window.close()

#Main
db_list=get_rs_list("master","select name from sys.databases",())
tab_list=[[],[]]

sg.theme('Dark Green 5')
menu_def=[['Databases(s)', db_list]]
layout=[
[[sg.Listbox(values=db_list[1], select_mode='extended',key='-dblist-', size=(30, 10),enable_events=True)],
[sg.Listbox(values=tab_list[1], select_mode='extended',key='-tablist-', size=(30, 20),enable_events=True)]]
]
win =sg.Window('Database Table CRUD tool',layout,size=(820, 640))
dbname=""
tabname=""
while True:
    event, data = win.read()
    print(event, data)
    if event in (sg.WIN_CLOSED,'Cancel','Exit'):
       break
    if event in ('-dblist-'):
       dbname=data['-dblist-'][0][0]
       tab_list=get_rs_list(dbname,"select name from sys.objects where type='U'",())
       win['-tablist-'].update(tab_list[1])
    if event in ('-tablist-'):
       tabname=data['-tablist-'][0][0]
       de_form(dbname,tabname)
win.close()

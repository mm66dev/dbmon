import PySimpleGUI as sg
import pymssql

def clear_de_form(window,fields):
    for f in fields:    
        window[f["name"]].update("")
def de_form(table_name):
    #print("table_name: {}".format(table_name))
    frame_layout=[]
    cn = pymssql.connect(server = "(local)\SQLEXPRESS",user="sa",password="*****",database="payroll")
    cur = cn.cursor(as_dict=True)
    frame_layout=[]
    fields=[]
    cur.execute("""select C.ORDINAL_POSITION cid,C.COLUMN_NAME name, iif((select 1 
	FROM INFORMATION_SCHEMA.TABLE_CONSTRAINTS T
	JOIN INFORMATION_SCHEMA.CONSTRAINT_COLUMN_USAGE CU 
	ON CU.CONSTRAINT_NAME=T.CONSTRAINT_NAME
	WHERE CU.TABLE_NAME=C.TABLE_NAME and CU.COLUMN_NAME=C.COLUMN_NAME) is null,0,1) as pk,
    COLUMNPROPERTY(object_id(C.TABLE_SCHEMA+'.'+C.TABLE_NAME), C.COLUMN_NAME, 'IsIdentity') id
    FROM INFORMATION_SCHEMA.COLUMNS C WHERE C.TABLE_NAME=%s""",(table_name,))
    c1=cur.fetchall()
    for row in c1:
        fields.append(row)
    pkf=[]
    idf=[]
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
            [sg.Frame(table_name, frame_layout, font='Any 12', title_color='blue',key="frame1")],
            [sg.Button("Save"),sg.Button("Find"), sg.Button("Delete"),sg.Button("Clear"),sg.Cancel()]
            ]
    window = sg.Window('{} - Table CRUD form'.format(table_name), layout,finalize=True,modal=True)
    found_flag=True
    while True:
        event, data = window.read()
        print(event, data)
        if event in (sg.WIN_CLOSED,'Cancel','Exit'):
            break
        if event in ("Find"):
            found_flag=False
            where_str=' and '.join(f["name"]+"=%s" for f in fields if f["cid"] in pkf)
            where_data=[ window[f["name"]].get() for f in fields if f["cid"] in pkf ]
            sql="select * from " + table_name + " where " + where_str
            cur.execute(sql,tuple(where_data) )
            c1=cur.fetchall()
            for row in c1:
                found_flag=True
                for f in fields:
                    window[f["name"]].update(row[f["name"]])
            if not found_flag:
                sg.popup("Record not found")
        elif event in ("Save"):
            try:
                where_str=' and '.join(f["name"]+"=%s" for f in fields if f["cid"] in pkf)
                where_data=[ window[f["name"]].get() for f in fields if f["cid"] in pkf ]
                if found_flag:
                    set_fields=','.join(f["name"] + "=%s" for f in fields if f["cid"] not in pkf + idf)
                    set_data=[window[f["name"]].get() for f in fields if f["cid"] not in pkf + idf]
                    sql="update " + table_name + " set " + set_fields + " where " + where_str
                    print(sql,set_data)
                    cur.execute(sql,tuple(set_data + set_data))
                    cn.commit()
                else:
                    insert_cols=",".join(f["name"] for f in fields if f["cid"] not in idf)
                    insert_parms=",".join("%s" for f in fields if f["cid"] not in idf)
                    insert_values=",".join("'" + window[f["name"]].get() + "'" for f in fields if f["cid"] not in idf)
                    insert_data=[window[f["name"]].get() for f in fields  if f["cid"] not in idf ]
                    sql="insert into " + table_name + " (" + insert_cols + ") values (" + insert_parms + ")"
                    print(sql,insert_data)
                    cur.execute(sql,tuple(insert_data))
                    cn.commit()
                    found_flag=False
                clear_de_form(window,fields)    
            except Exception as ex:
                sg.popup_ok("Error: {}".format(ex))
        elif event in ("Delete"):
            where_str=' and '.join(f["name"]+"=%s" for f in fields if f["cid"] in pkf)
            where_data=[ window[f["name"]].get() for f in fields if f["cid"] in pkf ]
            sql="delete from " + table_name + " where " + where_str
            cur.execute(sql,tuple(where_data))
            cn.commit()
            clear_de_form(window,fields)    
        elif event in ("Clear"):
            found_flag=False
            clear_de_form(window,fields)    
    cur.close()
    cn.close()

    window.close()

#Main
cn = pymssql.connect(server = "(local)\SQLEXPRESS",user="sa",password="*****",database="payroll")
cur = cn.cursor(as_dict=False)
table_list=[]
cur.execute("select name from sys.objects where type='U' order by 1")
c1=cur.fetchall()
for row in c1:
    table_list.append(row[0])
cur.close()
cn.close()
#print(table_list)
sg.theme('Dark Green 5')
menu_def=[['Table(s)', table_list]]
layout=[
    [ sg.Menu(menu_def) ]    ]
win =sg.Window('Database Table CRUD tool',layout,size=(820, 640))
while True:
    event, data = win.read()
    #print(event, data)
    if event in (sg.WIN_CLOSED,'Cancel','Exit'):
        break
    de_form(event)
win.close()

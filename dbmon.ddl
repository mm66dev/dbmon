CREATE TABLE cmd (
    cmd_id SERIAL PRIMARY KEY,
    name character varying(128),
    type character(3),
    cmd text,
    enabled boolean DEFAULT true,
    disabled boolean DEFAULT false,
    default_cron_id integer DEFAULT 1 NOT NULL
);
CREATE TABLE cron (
    cron_id SERIAL PRIMARY KEY,
    name character varying(128) NOT NULL,
    expr character varying(128) NOT NULL,
    enabled boolean DEFAULT true NOT NULL
);
CREATE TABLE ep (
    ep_id SERIAL PRIMARY KEY,
    name character varying(16),
    tag character varying(16),
    url character varying(4096) NOT NULL,
    enabled boolean DEFAULT true,
    disabled boolean DEFAULT false
);
CREATE TABLE ep_driver (
    epd_id character varying(16) NOT NULL,
    name character varying(128) NOT NULL,
    url_template character varying(4096) NOT NULL
);
CREATE TABLE epg (
    epg_id SERIAL PRIMARY KEY,
    name character varying(16)
);
CREATE TABLE epgi (
    epg_id integer NOT NULL,
    ep_id integer NOT NULL
);
CREATE TABLE epmg (
    epmg_id SERIAL PRIMARY KEY,
    name character varying(128) NOT NULL
);
CREATE TABLE epmgi (
    epmg_id integer NOT NULL,
    epg_id integer NOT NULL,
    cmd_id integer NOT NULL,
    cron_id integer NOT NULL
);
CREATE TABLE job_log (
    epmg_id integer,
    epg_id integer,
    cmd_id integer,
    cron_id integer,
    ep_id integer,
    start_ts text,
    end_ts text,
    result text
);
CREATE VIEW mp_info AS
 SELECT epmgi.epmg_id,
    epmgi.epg_id,
    epmgi.cmd_id,
    epmgi.cron_id,
    epgi.ep_id,
    ep.url,
    c.cmd
   FROM (((epmgi
     JOIN cmd c ON ((c.cmd_id = epmgi.cmd_id)))
     JOIN epgi ON ((epgi.epg_id = epmgi.epg_id)))
     JOIN ep ON ((ep.ep_id = epgi.ep_id)));

CREATE INDEX ix01_epmgi ON epmgi USING btree (cron_id);
ALTER TABLE ONLY epgi
    ADD CONSTRAINT epgi_ep_id_fkey FOREIGN KEY (ep_id) REFERENCES ep(ep_id);
ALTER TABLE ONLY epgi
    ADD CONSTRAINT epgi_epg_id_fkey FOREIGN KEY (epg_id) REFERENCES epg(epg_id);
ALTER TABLE ONLY epmgi
    ADD CONSTRAINT epmgi_cmd_id_fkey FOREIGN KEY (cmd_id) REFERENCES cmd(cmd_id);
ALTER TABLE ONLY epmgi
    ADD CONSTRAINT epmgi_cron_id_fkey FOREIGN KEY (cron_id) REFERENCES cron(cron_id);
ALTER TABLE ONLY epmgi
    ADD CONSTRAINT epmgi_epg_id_fkey FOREIGN KEY (epg_id) REFERENCES epg(epg_id);
ALTER TABLE ONLY epmgi
    ADD CONSTRAINT epmgi_epmg_id_fkey FOREIGN KEY (epmg_id) REFERENCES epmg(epmg_id);

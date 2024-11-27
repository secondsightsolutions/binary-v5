
CREATE TABLE atlas.scrubs (
    seq  bigserial,
    scid bigserial primary key,
    auth text not null,
    plcy text not null,
    name text not null,
    vers text not null,
    dscr text not null,
    hash text not null,
    host text not null,
    appl text not null,
    hdrs text not null,
    cmdl text not null,
    created timestamp not null default now(),
    ready   timestamp,
    started timestamp,
    done    timestamp
);

CREATE TABLE atlas.rebates (
    seq   bigserial,
    scid  bigint not null,
    rbid  bigint not null,
    indx  bigint not null default 0,
    rxn   text not null default '',
    hrxn  text not null default '',
    ndc   text not null default '',
    spid  text not null default '',
    prid  text not null default '',
    dos   text not null default '',
    stat  text not null default '',
    excl  text not null default '',
    errc  text not null default '',
    errm  text not null default '',
    spmt  text not null default '',
    col1  text not null default '',
    col2  text not null default '',
    col3  text not null default '',
    col4  text not null default '',
    col5  text not null default '',
    col6  text not null default '',
    col7  text not null default '',
    col8  text not null default '',
    col9  text not null default '',
    col10 text not null default '',
    col11 text not null default '',
    col12 text not null default '',
    col13 text not null default '',
    col14 text not null default '',
    col15 text not null default '',
    col16 text not null default '',
    col17 text not null default '',
    col18 text not null default '',
    col19 text not null default '',
    col20 text not null default '',
    col21 text not null default '',
    col22 text not null default '',
    col23 text not null default '',
    col24 text not null default '',
    col25 text not null default '',
    col26 text not null default '',
    col27 text not null default '',
    col28 text not null default '',
    col29 text not null default '',
    col30 text not null default '',
    col31 text not null default '',
    col32 text not null default '',
    col33 text not null default '',
    col34 text not null default '',
    col35 text not null default '',
    col36 text not null default '',
    col37 text not null default '',
    col38 text not null default '',
    col39 text not null default '',
    col40 text not null default '',
    col41 text not null default '',
    col42 text not null default '',
    col43 text not null default '',
    col44 text not null default '',
    col45 text not null default '',
    col46 text not null default '',
    col47 text not null default '',
    col48 text not null default '',
    col49 text not null default '',
    col50 text not null default '',
    CONSTRAINT rebates_pk PRIMARY KEY (scid, rbid)
);
CREATE INDEX ON atlas.rebates(seq);
CREATE INDEX ON atlas.rebates(scid);
CREATE INDEX ON atlas.rebates(scid, indx);

CREATE TABLE atlas.claim_uses (
    seq   bigserial,
    scid  bigint not null,
    shrt  text not null,
    excl  text not null default '',
    CONSTRAINT claim_uses_pk PRIMARY KEY (scid, shrt)
);
CREATE INDEX ON atlas.claim_uses(seq);
CREATE INDEX ON atlas.claim_uses(scid);
CREATE INDEX ON atlas.claim_uses(shrt);
CREATE INDEX ON atlas.claim_uses(scid, shrt);

CREATE TABLE atlas.rebate_meta (
    seq   bigserial,
    scid  bigint not null primary key,
    col1  text not null default '',
    col2  text not null default '',
    col3  text not null default '',
    col4  text not null default '',
    col5  text not null default '',
    col6  text not null default '',
    col7  text not null default '',
    col8  text not null default '',
    col9  text not null default '',
    col10 text not null default '',
    col11 text not null default '',
    col12 text not null default '',
    col13 text not null default '',
    col14 text not null default '',
    col15 text not null default '',
    col16 text not null default '',
    col17 text not null default '',
    col18 text not null default '',
    col19 text not null default '',
    col20 text not null default '',
    col21 text not null default '',
    col22 text not null default '',
    col23 text not null default '',
    col24 text not null default '',
    col25 text not null default '',
    col26 text not null default '',
    col27 text not null default '',
    col28 text not null default '',
    col29 text not null default '',
    col30 text not null default '',
    col31 text not null default '',
    col32 text not null default '',
    col33 text not null default '',
    col34 text not null default '',
    col35 text not null default '',
    col36 text not null default '',
    col37 text not null default '',
    col38 text not null default '',
    col39 text not null default '',
    col40 text not null default '',
    col41 text not null default '',
    col42 text not null default '',
    col43 text not null default '',
    col44 text not null default '',
    col45 text not null default '',
    col46 text not null default '',
    col47 text not null default '',
    col48 text not null default '',
    col49 text not null default '',
    col50 text not null default ''
);
CREATE INDEX ON atlas.rebate_meta(seq);
CREATE INDEX ON atlas.rebate_meta(scid);

CREATE TABLE atlas.rebate_claims (
    seq  bigserial,
    scid bigint not null,
    rbid bigint not null,
    shrt text not null,
    CONSTRAINT rebate_claims_pk PRIMARY KEY (scid, rbid, shrt)
);
CREATE INDEX ON atlas.rebate_claims(seq);

CREATE TABLE atlas.sync (
    pkey          integer primary key,
    claims        bigint not null default 0,
    scrubs        bigint not null default 0,
    rebates       bigint not null default 0,
    claim_uses    bigint not null default 0,
    rebate_meta   bigint not null default 0,
    rebate_claims bigint not null default 0
);

CREATE TABLE atlas.claims (
    shrt text primary key,
    i340 text not null,
    ndc  text not null,
    spid text not null,
    prid text not null default '',
    hrxn text not null,
    hfrx text not null default '',
    hdos text not null,
    hdop text not null,
    doc  bigint not null default 0,
    dos  bigint not null default 0,
    dop  bigint not null default 0,
    netw text not null,
    prnm text not null,
    chnm text not null default '',
    elig bool not null default true,
    susp bool not null default false,
    cnfm bool not null default true,
    qty  numeric not null default 0,
    manu text not null,
    ihph text array not null default '{}'
);
CREATE INDEX ON atlas.claims(doc);

CREATE TABLE atlas.proc (
	proc text not null,
	enb  bool not null default true,
	ver  int8 not null default 0,
	CONSTRAINT proc_pkey PRIMARY KEY (proc)
);

CREATE TABLE atlas.auth (
	auth text not null,
	enb  bool not null default true,
	kind text not null default 'pharmacy',
	CONSTRAINT auth_pkey PRIMARY KEY (auth)
);

CREATE TABLE atlas.proc_auth (
	proc text not null,
	auth text not null,
	CONSTRAINT proc_auth_auth_key UNIQUE (auth),
	CONSTRAINT proc_auth_pkey     PRIMARY KEY (proc, auth)
);

ALTER TABLE atlas.proc_auth ADD CONSTRAINT fk_proc_auth_auth FOREIGN KEY (auth) REFERENCES atlas.auth(auth);
ALTER TABLE atlas.proc_auth ADD CONSTRAINT fk_proc_auth_proc FOREIGN KEY (proc) REFERENCES atlas.proc(proc);

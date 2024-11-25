
CREATE TABLE titan.scrubs (
    manu text not null,
    scid bigint not null,
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
    rbt_total int not null default 0,
    rbt_valid int not null default 0,
    rbt_matched int not null default 0,
    rbt_nomatch int not null default 0,
    rbt_passed  int not null default 0,
    rbt_failed  int not null default 0,
    clm_total   int not null default 0,
    clm_valid   int not null default 0,
    clm_matched int not null default 0,
    clm_nomatch int not null default 0,
    clm_invalid int not null default 0,
    spi_exact   int not null default 0,
    spi_cross   int not null default 0,
    spi_stack   int not null default 0,
    spi_chain   int not null default 0,
    dos_equ_doc int not null default 0,
    dos_bef_doc int not null default 0,
    dos_equ_dof int not null default 0,
    dos_bef_dof int not null default 0,
    dos_aft_dof int not null default 0,
    CONSTRAINT PRIMARY KEY (manu, scid)
);

CREATE TABLE titan.rebates (
    manu text not null,
    scid bigint not null,
    rbid bigint not null,
    stat text not null default '',
    fprt text not null default '',
    CONSTRAINT PRIMARY KEY (manu, scid, rbid)
);
CREATE INDEX ON titan.rebates(manu);
CREATE INDEX ON titan.rebates(manu, scid);

CREATE TABLE titan.claim_uses (
    manu  text not null,
    scid  bigint not null,
    shrt  text not null,
    excl  text not null default '',
    CONSTRAINT claim_uses_pk PRIMARY KEY (manu, scid, shrt)
);
CREATE INDEX ON titan.claim_uses(scid);
CREATE INDEX ON titan.claim_uses(shrt);
CREATE INDEX ON titan.claim_uses(scid, shrt);

CREATE TABLE titan.rebate_meta (
    manu  text not null,
    scid  bigint not null,
    col1  text not null default '',
    col2  text not null default '',
    col50 text not null default '',
    CONSTRAINT rebate_meta_pk PRIMARY KEY (manu, scid)
);
CREATE INDEX ON titan.rebate_meta(scid);

CREATE TABLE titan.rebate_claims (
    manu text not null,
    scid bigint not null,
    rbid bigint not null,
    shrt text not null,
    CONSTRAINT rebate_claims_pk PRIMARY KEY (manu, scid, rbid, shrt)
);

-- PROVISIONING

CREATE TABLE titan.auth (
	auth text NOT NULL,
	enb  bool NOT NULL DEFAULT true,
	kind text NOT NULL DEFAULT 'pharmacy',
	CONSTRAINT auth_pkey PRIMARY KEY (auth)
);
GRANT ALL ON TABLE titan.auth TO titan;

CREATE TABLE titan.manu (
	manu text NOT NULL,
	enb  bool NOT NULL DEFAULT true,
	CONSTRAINT manu_pkey PRIMARY KEY (manu)
);
GRANT ALL ON TABLE titan.manu TO titan;

CREATE TABLE titan.manu_auth (
	manu text NOT NULL,
	auth text NOT NULL,
	enb  bool NOT NULL DEFAULT true,
	CONSTRAINT manu_auth_pkey PRIMARY KEY (manu, auth)
);
GRANT ALL ON TABLE titan.manu_auth TO titan;
ALTER TABLE titan.manu_auth ADD CONSTRAINT fk_manu_auth_auth FOREIGN KEY (auth) REFERENCES titan.auth(auth);
ALTER TABLE titan.manu_auth ADD CONSTRAINT fk_manu_auth_manu FOREIGN KEY (manu) REFERENCES titan.manu(manu);

CREATE TABLE titan.proc (
	proc text NOT NULL,
	enb  bool NOT NULL DEFAULT true,
	ver  int8 NOT NULL DEFAULT 0,
	CONSTRAINT proc_pkey PRIMARY KEY (prid)
);
GRANT ALL ON TABLE titan.proc TO titan;

CREATE TABLE titan.proc_auth (
	proc text NOT NULL,
	auth text NOT NULL,
	CONSTRAINT proc_auth_auth_key UNIQUE (auth),
	CONSTRAINT proc_auth_pkey PRIMARY KEY (proc, auth)
);
GRANT ALL ON TABLE titan.proc_auth TO titan;
ALTER TABLE titan.proc_auth ADD CONSTRAINT fk_proc_auth_auth FOREIGN KEY (auth) REFERENCES titan.auth(auth);
ALTER TABLE titan.proc_auth ADD CONSTRAINT fk_proc_auth_proc FOREIGN KEY (proc) REFERENCES titan.proc(proc);

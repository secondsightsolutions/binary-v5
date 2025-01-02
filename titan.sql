
CREATE TABLE titan.scrubs (
    manu text not null,
    scid bigint not null,
    auth text not null,
    plcy text not null,
    name text not null,
    kind text not null,
    vers text not null,
    dscr text not null,
    hash text not null,
    host text not null,
    appl text not null,
    hdrs text not null,
    cmdl text not null,
    crat bigint not null default 0,
    rdat bigint not null default 0,
    srat bigint not null default 0,
    dnat bigint not null default 0,
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
    seq         bigint not null,

    CONSTRAINT scrubs_pk PRIMARY KEY (manu, scid)
);

CREATE TABLE titan.rebates (
    manu text not null,
    scid bigint not null,
    rbid bigint not null,
    stat text not null default '',
    fprt text not null default '',
    seq  bigint not null,
    CONSTRAINT rebates_pk PRIMARY KEY (scid, rbid, manu)
);
CREATE INDEX ON titan.rebates(manu);
CREATE INDEX ON titan.rebates(manu, scid);

CREATE TABLE titan.claim_uses (
    manu  text not null,
    scid  bigint not null,
    shrt  text not null,
    excl  text not null default '',
    seq   bigint not null,
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
    seq  bigint not null,
    CONSTRAINT rebate_claims_pk PRIMARY KEY (manu, scid, rbid, shrt)
);

-- PROVISIONING

CREATE TABLE titan.auth (
    manu text not null,
	proc text not null,
    auth text not null,
    kind text not null default 'pharmacy',
	ver  int8 not null default 0,
    enb  bool not null default true,
	CONSTRAINT auth_pkey PRIMARY KEY (manu, proc, auth, kind)
);

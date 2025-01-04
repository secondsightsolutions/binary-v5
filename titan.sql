
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
    crat timestamp not null, -- created
    rdat timestamp, -- ready
    srat timestamp, -- started
    dnat timestamp, -- done
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

CREATE TABLE titan.rebate_claims (
    manu text not null,
    scid bigint not null,
    rbid bigint not null,
    shrt text not null,
    seq  bigint not null,
    CONSTRAINT rebate_claims_pk PRIMARY KEY (manu, scid, rbid, shrt)
);

CREATE TABLE titan.claims (
    manu text not null,
    shrt text not null,
    i340 text not null,
    ndc  text not null,
    spid text not null,
    prid text not null default '',
    hrxn text not null,
    hfrx text not null default '',
    hdos text not null,
    hdop text not null,
    doc  timestamp,
    dos  timestamp,
    dop  timestamp,
    netw text not null,
    prnm text not null,
    chnm text not null default '',
    elig bool not null default true,
    susp bool not null default false,
    cnfm bool not null default true,
    qty  numeric not null default 0,
    ihph text not null default '',
    seq  bigint not null,
    CONSTRAINT claims_pk PRIMARY KEY(manu, shrt)
);
CREATE INDEX ON titan.claims(manu);
CREATE INDEX ON titan.claims(manu, doc);

CREATE TABLE titan.auth (
    manu text not null,
	proc text not null,
    auth text not null,
    kind text not null default 'pharmacy',
	ver  int8 not null default 0,
    enb  bool not null default true,
	CONSTRAINT auth_pkey PRIMARY KEY (manu, proc, auth, kind)
);

CREATE TABLE titan.entities (
    i340 text   not null,
    strt date,
    term date,
    stat text   not null default '',
    seq  bigint not null primary key
);
CREATE INDEX ON titan.entities(seq);

CREATE TABLE titan.pharmacies (
    i340 text not null,
    phid text not null,
    ncps text not null default '',
    npis text not null default '',
    deas text not null default '',
    chnm text not null default '',
    stat text not null default '',
    seq  bigint not null primary key
);

CREATE TABLE titan.ndcs (
    manu text not null,
    ndc  text not null,
    name text not null,
    netw text not null default 'retail',
    seq  bigint not null,
    CONSTRAINT ndcs_pk PRIMARY KEY (manu, ndc)
);
CREATE INDEX ON titan.ndcs(manu);
CREATE INDEX ON titan.ndcs(seq);

CREATE TABLE titan.spis (
    ncp  text not null primary key,
    npi  text not null default '',
    dea  text not null default '',
    sto  text not null default '',
    nam  text not null default '',
    lbn  text not null default '',
    chn  text not null default '',
    cde  text not null default '',
    seq  bigint not null
);
CREATE INDEX ON titan.spis(seq);

CREATE TABLE titan.desigs (
    manu text not null,
    i340 text not null,
    phid text not null,
    netw text not null default 'retail',
    flag text not null default '',
    hin  text not null default '',
    assg boolean not null default true,
    term boolean not null default false,
    excl boolean not null default false,
    xdat timestamp,
    dlat timestamp,
    xsat timestamp,
    crat timestamp,
    cpat timestamp,
    seq  bigint not null,
    CONSTRAINT desigs_pk PRIMARY KEY (manu, i340, phid)
);
CREATE INDEX ON titan.desigs(manu);

CREATE TABLE titan.ldns (
    manu text not null,
    netw text not null,
    phid text not null,
    assg boolean not null default true,
    term boolean not null default false,
    seq  bigint not null,
    CONSTRAINT ldns_pk PRIMARY KEY (manu, netw, phid)
);
CREATE INDEX ON titan.ldns(manu);

CREATE TABLE titan.esp1 (
    manu text   not null,
    spid text   not null,
    ndc  text   not null,
    strt date,
    term date,
    CONSTRAINT esp1_pk PRIMARY KEY (manu, spid, ndc)
);

CREATE TABLE titan.eligibility (
    seq  bigint not null primary key,
    manu text   not null,
    i340 text   not null,
    phid text   not null,
    netw text   not null default 'retail',
    strt timestamp,
    term timestamp
);
CREATE INDEX ON titan.eligibility(manu);


DROP TABLE IF EXISTS atlas.scrub_rebates_claims;
DROP TABLE IF EXISTS atlas.scrub_rebates;
DROP TABLE IF EXISTS atlas.scrub_claims;
DROP TABLE IF EXISTS atlas.metrics;
DROP TABLE IF EXISTS atlas.scrubs;
DROP TABLE IF EXISTS atlas.rebates;
DROP TABLE IF EXISTS atlas.claims;
DROP TABLE IF EXISTS atlas.invoice_cols;
DROP TABLE IF EXISTS atlas.invoice_rows;
DROP TABLE IF EXISTS atlas.invoices;
DROP TABLE IF EXISTS atlas.auth;
DROP TABLE IF EXISTS atlas.sync;
DROP TABLE IF EXISTS atlas.commands;
DROP TABLE IF EXISTS atlas.test_invoices;
DROP TABLE IF EXISTS atlas.test_rebates;
DROP TABLE IF EXISTS atlas.test_claims;
DROP TABLE IF EXISTS atlas.test_desigs;
DROP TABLE IF EXISTS atlas.test_eligibilities;
DROP TABLE IF EXISTS atlas.test_entities;
DROP TABLE IF EXISTS atlas.test_esp1;
DROP TABLE IF EXISTS atlas.test_ldns;
DROP TABLE IF EXISTS atlas.test_ndcs;
DROP TABLE IF EXISTS atlas.test_pharmacies;
DROP TABLE IF EXISTS atlas.test_spis;

CREATE TABLE atlas.sync (
    commands        bigint not null default -1,
    scrubs          bigint not null default -1,
    scrub_rebates   bigint not null default -1,
    scrub_claims    bigint not null default -1,
    scrub_reb_clms  bigint not null default -1,
    metrics         bigint not null default -1,
    pkey            bigint not null default 1 primary key
);

CREATE TABLE atlas.commands (
    cmid bigserial primary key,
    comd text not null,
    manu text not null,
    name text not null,
    auth text not null,
    vers text not null,
    kind text not null default '',
    xou  text not null,
    dscr text not null,
    hash text not null,
    netw text not null default '',
    host text not null default '',
    addr text not null default '',
    cwd  text not null default '',
    "user" text not null default '',
    cmdl text not null,
    rslt text not null default '',
    crat timestamp with time zone not null default now(), -- created
    seq  bigserial
);
CREATE INDEX ON atlas.commands(seq);

CREATE TABLE atlas.invoices (
    manu text not null,
	ivid bigserial primary key,
    cmid bigint not null references atlas.commands(cmid),
    file text not null default '',
    crat timestamp with time zone not null default now()
);

CREATE TABLE atlas.invoice_cols (
    manu text not null,
    ivid bigint not null,
    indx int2 not null,
    name text not null,
    CONSTRAINT invoice_cols_pk PRIMARY KEY (ivid, indx),
    FOREIGN KEY (ivid) REFERENCES atlas.invoices(ivid)
);
CREATE INDEX ON atlas.invoice_cols(ivid);
CREATE UNIQUE INDEX ON atlas.invoice_cols(ivid, name);

CREATE TABLE atlas.invoice_rows (
    manu text not null,
    ivid bigint not null,
    rbid int8 not null,
    data text not null,
    CONSTRAINT invoice_data_pk PRIMARY KEY (ivid, rbid),
    FOREIGN KEY (ivid) REFERENCES atlas.invoices(ivid)
);
CREATE INDEX ON atlas.invoice_rows(ivid);

CREATE TABLE atlas.rebates (
    manu text not null,
    ivid bigint not null references atlas.invoices(ivid),
    rbid bigint not null,
    ndc  text not null default '',
    rxn  text not null default '',
    hrxn text not null default '',
    spid text not null default '',
    hspd text not null default '',
    prid text not null default '',
    hprd text not null default '',
    dos  timestamp with time zone,
    hdos text not null default '',
    CONSTRAINT rebates_pk PRIMARY KEY (ivid, rbid)
);
CREATE INDEX ON atlas.rebates(ivid);
CREATE INDEX ON atlas.rebates(rbid);

CREATE TABLE atlas.scrubs (
    manu text not null,
    scid bigserial primary key,
    ivid bigint not null references atlas.invoices(ivid),
    cmid bigint not null references atlas.commands(cmid),
    plcy text not null,
    hdrs text not null,
    crat timestamp with time zone not null default now(), -- created
    rdat timestamp with time zone, -- ready
    srat timestamp with time zone, -- started
    dnat timestamp with time zone, -- done
    test text not null default '',
    seq  bigserial
);
CREATE INDEX ON atlas.scrubs(seq);

CREATE TABLE atlas.metrics (
    scid bigint not null primary key references atlas.scrubs(scid),
    manu                text not null,
    rbt_total           integer not null default 0,
    rbt_matched         integer not null default 0,
    rbt_nomatch         integer not null default 0,
    rbt_invalid         integer not null default 0,
    rbt_passed          integer not null default 0,
    rbt_failed          integer not null default 0,
    clm_total           integer not null default 0,
    clm_valid           integer not null default 0,
    clm_matched         integer not null default 0,
    clm_nomatch         integer not null default 0,
    clm_invalid         integer not null default 0,
    spi_exact           integer not null default 0,
    spi_cross           integer not null default 0,
    spi_stack           integer not null default 0,
    spi_chain           integer not null default 0,
    dos_equ_doc         integer not null default 0,
    dos_bef_doc         integer not null default 0,
    dos_aft_doc         integer not null default 0,
    dos_equ_dof         integer not null default 0,
    dos_bef_dof         integer not null default 0,
    dos_aft_dof         integer not null default 0,
    dos_range_pass      integer not null default 0,
    dos_range_fail      integer not null default 0,
    dof_range_pass      integer not null default 0,
    dof_range_fail      integer not null default 0,
    doc_range_pass      integer not null default 0,
    doc_range_fail      integer not null default 0,
    r_no_match_rx       integer not null default 0,
    r_no_match_spi      integer not null default 0,
    r_no_match_ndc      integer not null default 0,
    r_clm_used          integer not null default 0,
    r_dos_rbt_aft_clm   integer not null default 0,
    r_dos_clm_aft_rbt   integer not null default 0,
    r_old_rbt_new_clm   integer not null default 0,
    r_new_rbt_old_clm   integer not null default 0,
    r_clm_not_cnfm      integer not null default 0,
    r_phm_not_desg      integer not null default 0,
    r_inv_desg_type     integer not null default 0,
    r_wrong_network     integer not null default 0,
    r_not_elig_at_sub   integer not null default 0,
    seq                 bigserial
);

CREATE TABLE atlas.claims (
    manu text not null,
    clid text primary key,
    i340 text not null default '',
    ndc  text not null default '',
    spid text not null default '',
    prid text not null default '',
    hrxn text not null default '',
    hfrx text not null default '',
    hdos text not null default '',
    hdop text not null default '',
    doc  time with time zone,
    dos  timestamp with time zone,
    dop  timestamp with time zone,
    netw text not null default '',
    prnm text not null default '',
    chnm text not null default '',
    elig bool not null default true,
    susp bool not null default false,
    cnfm bool not null default true,
    qty  numeric not null default 0,
    ihph text not null default '',
    seq  bigint not null
);
CREATE INDEX ON atlas.claims(manu);
CREATE INDEX ON atlas.claims(seq);

CREATE TABLE atlas.scrub_rebates (
    manu text not null,
    scid bigint not null, -- references atlas.scrubs(scid),
    ivid bigint not null references atlas.invoices(ivid),
    rbid bigint not null,
    indx bigint not null default 0,
    stat text not null default '',
    excl text not null default '',
    spmt text not null default '',
    errc text not null default '',
    errm text not null default '',
    seq  bigserial,
    CONSTRAINT scrub_rebates_pk PRIMARY KEY (scid, rbid),
    FOREIGN KEY (ivid, rbid) REFERENCES atlas.rebates (ivid, rbid)
);

CREATE TABLE atlas.scrub_claims (
    manu  text not null,
    scid bigint not null, -- references atlas.scrubs(scid),
    clid text   not null references atlas.claims(clid),
    excl text not null default '',
    seq  bigserial,
    CONSTRAINT scrub_claims_pk PRIMARY KEY (scid, clid)
);
CREATE INDEX ON atlas.scrub_claims(seq);

CREATE TABLE atlas.scrub_rebates_claims (
    manu  text not null,
    scid bigint not null, -- references atlas.scrubs(scid),
    ivid bigint not null references atlas.invoices(ivid),
    rbid bigint not null,
    clid text   not null references atlas.claims,
    seq  bigserial,
    CONSTRAINT scrub_rebates_claims_pk PRIMARY KEY (scid, rbid, clid),
    FOREIGN KEY (ivid, rbid) REFERENCES atlas.rebates (ivid, rbid) --,
    --FOREIGN KEY (scid, rbid) REFERENCES atlas.scrub_rebates (scid, rbid),
    --FOREIGN KEY (ivid, clid) REFERENCES atlas.scrub_claims (scid, clid)
);

CREATE TABLE atlas.auth (
    manu text not null,
	proc text not null,
    auth text not null,
    kind text not null default 'pharmacy',
	vers int8 not null default 0,
    enb  bool not null default true,
	CONSTRAINT auth_pkey PRIMARY KEY (manu, proc, auth, kind)
);

-- For tests
CREATE TABLE atlas.test_invoices (
    manu text not null,
    ivid bigserial,
    crat timestamp with time zone not null default now(),
    CONSTRAINT test_invoices_pk PRIMARY KEY (manu, ivid)
);

CREATE TABLE atlas.test_rebates (
    manu  text not null,
    test  text not null,
    ivid  bigint not null,
    rbid  bigserial,
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
    CONSTRAINT test_rebates_pk PRIMARY KEY (manu, test, ivid, rbid)
);
CREATE INDEX ON atlas.test_rebates(test);

CREATE TABLE atlas.test_claims (
    manu text not null,
    test text not null,
    clid text not null,
    i340 text not null,
    ndc  text not null,
    spid text not null,
    prid text not null default '',
    hrxn text not null,
    hfrx text not null default '',
    hdos text not null,
    hdop text not null,
    doc  timestamp with time zone not null default now(),
    dos  timestamp with time zone not null default now(),
    dop  timestamp with time zone not null default now(),
    netw text not null,
    prnm text not null,
    chnm text not null default '',
    elig bool not null default true,
    susp bool not null default false,
    cnfm bool not null default true,
    qty  numeric not null default 0,
    ihph text not null default '',
    CONSTRAINT test_claims_pk PRIMARY KEY (manu, test, clid)
);
CREATE INDEX ON atlas.test_claims(test);

CREATE TABLE atlas.test_entities (
    manu text   not null,
    test text   not null,
    i340 text   not null,
    strt date not null,
    term date,
    dop  timestamp with time zone,
    stat text   not null default '',
    CONSTRAINT test_entities_pk PRIMARY KEY (manu, test, i340)
);
CREATE INDEX ON atlas.test_entities(test);

CREATE TABLE atlas.test_pharmacies (
    test text not null,
    i340 text not null,
    phid text not null,
    ncps text not null default '',
    npis text not null default '',
    deas text not null default '',
    chnm text not null default '',
    stat text not null default '',
    CONSTRAINT test_pharmacies_pk PRIMARY KEY (test, i340, phid)
);
CREATE INDEX ON atlas.test_pharmacies(test);

CREATE TABLE atlas.test_ndcs (
    manu text not null,
    test text not null,
    ndc  text not null,
    name text not null,
    netw text not null default 'retail',
    CONSTRAINT test_ndcs_pk PRIMARY KEY (manu, test, ndc)
);
CREATE INDEX ON atlas.test_ndcs(test);

CREATE TABLE atlas.test_spis (
    test text not null,
    ncp  text not null,
    npi  text not null default '',
    dea  text not null default '',
    sto  text not null default '',
    nam  text not null default '',
    lbn  text not null default '',
    chn  text not null default '',
    cde  text not null default '',
    CONSTRAINT test_spis_pk PRIMARY KEY (test, ncp)
);
CREATE INDEX ON atlas.test_claims(test);

CREATE TABLE atlas.test_desigs (
    manu text not null,
    test text not null,
    i340 text not null,
    phid text not null,
    netw text not null default 'retail',
    flag text not null default '',
    hin  text not null default '',
    assg boolean not null default true,
    term boolean not null default false,
    excl boolean not null default false,
    xdat timestamp with time zone,
    dlat timestamp with time zone,
    xsat timestamp with time zone,
    crat timestamp with time zone not null,
    cpat timestamp with time zone,
    CONSTRAINT test_desigs_pk PRIMARY KEY (manu, test, i340, phid)
);
CREATE INDEX ON atlas.test_claims(test);

CREATE TABLE atlas.test_ldns (
    manu text not null,
    test text not null,
    netw text not null,
    phid text not null,
    assg boolean not null default true,
    term boolean not null default false,
    CONSTRAINT test_ldns_pk PRIMARY KEY (manu, test, netw, phid)
);
CREATE INDEX ON atlas.test_ldns(test);

CREATE TABLE atlas.test_esp1 (
    test text   not null,
    ndc  text   not null,
    strt timestamp with time zone not null,
    term timestamp with time zone,
    CONSTRAINT test_esp1_pk PRIMARY KEY (test, ndc)
);
CREATE INDEX ON atlas.test_esp1(test, ndc);

CREATE TABLE atlas.test_eligibilities (
    manu text   not null,
    test text   not null,
    i340 text   not null,
    phid text   not null,
    netw text   not null default 'retail',
    strt timestamp with time zone not null,
    term timestamp with time zone,
    CONSTRAINT test_eligibilities_pk PRIMARY KEY (manu, test, i340, phid)
);
CREATE INDEX ON atlas.test_eligibilities(test);

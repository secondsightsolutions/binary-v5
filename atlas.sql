
DROP TABLE IF EXISTS atlas.scrub_matches;
DROP TABLE IF EXISTS atlas.scrub_attempts;
DROP TABLE IF EXISTS atlas.scrub_rebates;
DROP TABLE IF EXISTS atlas.scrub_claims;
DROP TABLE IF EXISTS atlas.metrics;
DROP TABLE IF EXISTS atlas.scrubs;
DROP TABLE IF EXISTS atlas.rebates;
DROP TABLE IF EXISTS atlas.claims;
DROP TABLE IF EXISTS atlas.invoice_cols;
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
	ivid bigserial,
    cmid bigint not null,
    file text not null default '',
    crat timestamp with time zone not null default now(),
    CONSTRAINT invoices_pk PRIMARY KEY (manu, ivid)
);

CREATE TABLE atlas.invoice_cols (
    manu text not null,
    ivid bigint not null,
    indx int2 not null,
    name text not null,
    CONSTRAINT invoice_cols_pk PRIMARY KEY (manu, ivid, indx),
    FOREIGN KEY (manu, ivid) REFERENCES atlas.invoices(manu, ivid)
);
CREATE INDEX ON atlas.invoice_cols(manu, ivid);
CREATE UNIQUE INDEX ON atlas.invoice_cols(manu, ivid, name);

--CREATE TABLE atlas.invoice_rows (
 --   manu text not null,
 --   ivid bigint not null,
 --   rbid int8 not null,
 --   data text not null,
 --   CONSTRAINT invoice_rows_pk PRIMARY KEY (manu, ivid, rbid),
 --   FOREIGN KEY (manu, ivid) REFERENCES atlas.invoices(manu, ivid)
--);
--CREATE INDEX ON atlas.invoice_rows(manu, ivid);

CREATE TABLE atlas.rebates (
    manu text not null,
    ivid bigint not null,
    rbid bigint not null,
    ndc  text not null default '',
    rxn  text not null default '',
    hrxn text not null default '',
    spid text not null default '',
    prid text not null default '',
    dos  timestamp with time zone,
    hdos text not null default '',
    data text not null default '',
    CONSTRAINT rebates_pk PRIMARY KEY (manu, ivid, rbid),
    FOREIGN KEY (manu, ivid) REFERENCES atlas.invoices(manu, ivid)
);
CREATE INDEX ON atlas.rebates(manu, ivid);
CREATE INDEX ON atlas.rebates(manu, rbid);

CREATE TABLE atlas.scrubs (
    manu text not null,
    scid bigserial,
    ivid bigint not null,
    cmid bigint not null,
    plcy text not null,
    prof text not null default '',
    cust text not null default '',
    crat timestamp with time zone not null default now(), -- created
    rdat timestamp with time zone, -- ready
    srat timestamp with time zone, -- started
    dnat timestamp with time zone, -- done
    test text not null default '',
    seq  bigserial,
    CONSTRAINT scrubs_pk PRIMARY KEY (manu, scid),
    FOREIGN KEY (manu, ivid) REFERENCES atlas.invoices(manu, ivid)
);
CREATE INDEX ON atlas.scrubs(seq);

CREATE TABLE atlas.metrics (
    manu                text not null,
    scid                bigint not null,
    ivid                bigint not null,
    rbt_total           integer not null default 0,
    rbt_matched         integer not null default 0,
    rbt_nomatch         integer not null default 0,
    rbt_valid           integer not null default 0,
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
    load_rebates        integer not null default 0,
    prep_claims         integer not null default 0,
    pull_rebates        integer not null default 0,
    work_rebates        integer not null default 0,
    save_rebates        integer not null default 0,
    save_scrub_rebates  integer not null default 0,
    save_scrub_matches  integer not null default 0,
    save_scrub_attempts integer not null default 0,
    save_scrub_claims   integer not null default 0,
    seq                 bigserial,
    CONSTRAINT metrics_pk PRIMARY KEY (manu, scid),
    FOREIGN KEY (manu, scid) REFERENCES atlas.scrubs(manu, scid)
);

CREATE TABLE atlas.claims (
    manu text not null,
    clid text not null,
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
    dupl bool not null default false,
    elig bool not null default true,
    susp bool not null default false,
    cnfm bool not null default true,
    qty  numeric not null default 0,
    ihph text not null default '',
    seq  bigint not null,
    CONSTRAINT claims_pk PRIMARY KEY (manu, clid)
);
CREATE INDEX ON atlas.claims(manu);
CREATE INDEX ON atlas.claims(manu, seq);

CREATE TABLE atlas.scrub_rebates (
    manu text not null,
    scid bigint not null,
    ivid bigint not null,
    rbid bigint not null,
    stat text not null default '',
    spmt text not null default '',
    fprt text not null default '',
    excl text not null default '',
    errc text not null default '',
    errm text not null default '',
    cust text not null default '',
    seq  bigserial,
    CONSTRAINT scrub_rebates_pk PRIMARY KEY (manu, scid, rbid),
    FOREIGN KEY (manu, ivid, rbid) REFERENCES atlas.rebates (manu, ivid, rbid),
    FOREIGN KEY (manu, scid)       REFERENCES atlas.scrubs  (manu, scid)
);
CREATE INDEX ON atlas.scrub_rebates(manu, ivid, scid);

CREATE TABLE atlas.scrub_claims (
    manu text   not null,
    scid bigint not null, 
    clid text   not null,
    excl text   not null default '',
    seq  bigserial,
    CONSTRAINT scrub_claims_pk PRIMARY KEY (manu, scid, clid)
);
CREATE INDEX ON atlas.scrub_claims(seq);

CREATE TABLE atlas.scrub_matches (
    manu text not null,
    scid bigint not null,
    ivid bigint not null,
    rbid bigint not null,
    clid text   not null,
    seq  bigserial,
    CONSTRAINT scrub_matches_pk PRIMARY KEY (manu, scid, rbid, clid),
    FOREIGN KEY (manu, ivid, rbid) REFERENCES atlas.rebates (manu, ivid, rbid),
    FOREIGN KEY (manu, scid)       REFERENCES atlas.scrubs  (manu, scid)
);
CREATE TABLE atlas.scrub_attempts (
    manu text not null,
    scid bigint not null,
    ivid bigint not null,
    rbid bigint not null,
    clid text   not null,
    excl text   not null,
    seq  bigserial,
    CONSTRAINT scrub_attempts_pk PRIMARY KEY (manu, scid, rbid, clid),
    FOREIGN KEY (manu, ivid, rbid) REFERENCES atlas.rebates (manu, ivid, rbid),
    FOREIGN KEY (manu, scid)       REFERENCES atlas.scrubs  (manu, scid)
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


CREATE TABLE atlas.scrubs (
    manu text not null,
    scid bigserial primary key,
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
    crat bigint not null default 0, -- created
    rdat bigint not null default 0, -- ready
    srat bigint not null default 0, -- started
    dnat bigint not null default 0, -- done
    test text not null default '',
    seq  bigserial
);

CREATE TABLE atlas.rebates (
    manu text not null,
    scid bigint not null,
    rbid bigserial,
    indx bigint not null default 0,
    rxn  text not null default '',
    hrxn text not null default '',
    ndc  text not null default '',
    spid text not null default '',
    prid text not null default '',
    dos  text not null default '',
    stat text not null default '',
    excl text not null default '',
    errc text not null default '',
    errm text not null default '',
    spmt text not null default '',
    seq  bigserial,
    CONSTRAINT rebates_pk PRIMARY KEY (scid, rbid)
);
CREATE INDEX ON atlas.rebates(seq);
CREATE INDEX ON atlas.rebates(scid);
CREATE INDEX ON atlas.rebates(scid, indx);

CREATE TABLE atlas.claim_uses (
    manu text not null,
    scid  bigint not null,
    shrt  text not null,
    excl  text not null default '',
    seq   bigserial,
    CONSTRAINT claim_uses_pk PRIMARY KEY (scid, shrt)
);
CREATE INDEX ON atlas.claim_uses(seq);
CREATE INDEX ON atlas.claim_uses(scid);
CREATE INDEX ON atlas.claim_uses(shrt);
CREATE INDEX ON atlas.claim_uses(scid, shrt);

CREATE TABLE atlas.rebate_meta (
    manu text not null,
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
    manu text not null,
    scid bigint not null,
    rbid bigint not null,
    shrt text not null,
    seq  bigserial,
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
    manu text not null,
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
    ihph text array not null default '{}'
);
CREATE INDEX ON atlas.claims(doc);

CREATE TABLE atlas.auth (
    manu text not null,
	proc text not null,
    auth text not null,
    kind text not null default 'pharmacy',
	ver  int8 not null default 0,
    enb  bool not null default true,
	CONSTRAINT auth_pkey PRIMARY KEY (manu, proc, auth, kind)
);

-- For tests
CREATE TABLE atlas.test_rebates (
    manu  text not null,
    test  text not null,
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
    CONSTRAINT test_rebates_pk PRIMARY KEY (manu, test, scid, rbid)
);
CREATE INDEX ON atlas.test_rebates(test);

CREATE TABLE atlas.test_claims (
    manu text not null,
    test text not null,
    shrt text not null,
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
    ihph text array not null default '{}',
    CONSTRAINT test_claims_pk PRIMARY KEY (manu, test, shrt)
);
CREATE INDEX ON atlas.test_claims(test);

CREATE TABLE atlas.test_entities (
    manu text   not null,
    test text   not null,
    i340 text   not null,
    strt bigint not null default 0,
    term bigint not null default 0,
    dop  bigint not null default 0,
    stat text   not null default '',
    CONSTRAINT test_entities_pk PRIMARY KEY (manu, test, i340)
);
CREATE INDEX ON atlas.test_entities(test);

CREATE TABLE atlas.test_pharmacies (
    manu text not null,
    test text not null,
    i340 text not null,
    phid text not null,
    ncps text array not null default '{}',
    npis text array not null default '{}',
    deas text array not null default '{}',
    chnm text not null default '',
    stat text not null default '',
    CONSTRAINT test_pharmacies_pk PRIMARY KEY (manu, test, i340, phid)
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
    manu text not null,
    test text not null,
    ncp  text not null,
    npi  text not null default '',
    dea  text not null default '',
    sto  text not null default '',
    nam  text not null default '',
    lbn  text not null default '',
    chn  text not null default '',
    cde  text not null default '',
    CONSTRAINT test_spis_pk PRIMARY KEY (manu, test, ncp)
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
    xdat bigint not null default 0,
    dlat bigint not null default 0,
    xsat bigint not null default 0,
    crat bigint not null default 0,
    cpat bigint not null default 0,
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
    manu text not null,
    test text   not null,
    ndc  text   not null,
    strt bigint not null default 0,
    term bigint not null default 0,
    CONSTRAINT test_esp1_pk PRIMARY KEY (manu, test, ndc)
);
CREATE INDEX ON atlas.test_esp1(test, ndc);

CREATE TABLE atlas.test_eligibilities (
    manu text   not null,
    test text   not null,
    i340 text   not null,
    phid text   not null,
    netw text   not null default 'retail',
    strt bigint not null default 0,
    term bigint not null default 0,
    CONSTRAINT test_eligibilities_pk PRIMARY KEY (manu, test, i340, phid)
);
CREATE INDEX ON atlas.test_eligibilities(test);

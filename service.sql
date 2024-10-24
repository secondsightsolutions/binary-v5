
CREATE TABLE v5.scrubs (
    manu text not null,
    scid bigint not null,
    CONSTRAINT PRIMARY KEY (manu, scid)
);

CREATE TABLE v5.scrub_rebates (
    manu text not null,
    scid bigint not null,
    rbid bigint not null,
    stat text not null default '',
    fprt text not null default '',
    CONSTRAINT PRIMARY KEY (manu, scid, rbid)
);
CREATE INDEX ON v5.scrub_rebates(manu);
CREATE INDEX ON v5.scrub_rebates(manu, scid);

CREATE TABLE v5.scrub_claims (
    manu text not null,
    scid bigint not null,
    shrt text not null,
    excl text not null default '',
    CONSTRAINT PRIMARY KEY (manu, scid, shrt)
);
CREATE INDEX ON v5.scrub_claims(manu, scid);
CREATE INDEX ON v5.scrub_claims(manu, shrt);
CREATE INDEX ON v5.scrub_claims(manu, scid, shrt);

CREATE TABLE v5.scrub_rebate_meta (
    manu text not null,
    scid  bigint not null primary key,
    col1  text not null default '',
    col2  text not null default '',
    col50 text not null default '',
    CONSTRAINT PRIMARY KEY (manu, scid)
);
CREATE INDEX ON v5.scrub_rebate_meta(manu, scid);

CREATE TABLE v5.scrub_rebate_claims (
    manu text not null,
    scid bigint not null,
    rbid bigint not null,
    shrt text not null,
    CONSTRAINT PRIMARY KEY (manu, scid, rbid, shrt)
);
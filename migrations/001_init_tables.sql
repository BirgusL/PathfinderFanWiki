CREATE TABLE Spells_EN
(
    id            serial       not null unique,
    Name          text         not null,
    Description   text
);

CREATE TABLE Spells_Info
(
    id                  serial       not null unique,
    School              text,
    Target              text,
    Duration            text,
    Level               integer,
    Saving_Throw        text,
    Action_Type         text,  
    Spell_Resist        text,
    Image               text,
    TypeJSON            text,
    DescriptorsJSON     text,
    RequirementsJSON    text,
    MetamagicJSON       text,
    CraftingJSON        text,
);

CREATE TABLE Visit_Info
(
    id                  serial       not null unique,
    Ip                  text,
    User_Agent          text,
    Referer             text,
    Path                text,
    Visit_Time          text,
    Session_id          text,  
);
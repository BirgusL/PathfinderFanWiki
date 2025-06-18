INSERT INTO Spells_EN (Name, Description) VALUES
    ('Magic Missile', 'Creates three glowing darts of magical force.'),
    ('Fireball', 'A fiery explosion detonates at the target location.');

INSERT INTO Spells_Info (spell_id, School, Target, Duration, Level, Saving_Throw, Action_Type, Spell_Resist, Image, TypeJSON, DescriptorsJSON, RequirementsJSON, MetamagicJSON, CraftingJSON) VALUES
    (1, 'Evocation', 'One enemy creature within long range', '1 round + 1 round per three levels', 2, 'No', 'Standard Action', 'Not affected by spell resistance', 'AcidArrow', '["Dmg(RTA)"]', '["Acid"]', '["Wizard/Sorcerer","Earth Domain","Magus","Water Domain","Bloodrager","Elements patron","Water Domain","Demon","Magic Deceiver"]', '["Empower","Maximize","Quicken","Extend","Heighten"]', '["Rainbow Quartz","Forked Tongue"]');
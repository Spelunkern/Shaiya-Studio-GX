package sdata

// SkillColumnNames and SkillColumnTags are the skill columns of the lists
// (g_skillColumnNames / g_skillColumnTags). The lists index them from 8
// ("Fighter" is entry 8); the reconstructed C++ table lacks the first 8
// entries, which no list uses, so they are empty here.
var SkillColumnNames = []string{
	"", "", "", "", "", "", "", "",
	"Fighter", "Defenser", "Ranger", "Archer", "Mage", "Priest",
	"PrevSkill", "ReqLevel", "Grow", "SkillPoint", "TypeShow", "TypeAttack",
	"TypeEffect", "TypeDetail", "Need Wp 1", "Need Wp 2", "Need Wp 3", "Need Wp 4",
	"Need Wp 5", "Need Wp 6", "Need Wp 7", "Need Wp 8", "Need Wp 9", "Need Wp 10",
	"Need Wp 11", "Need Wp 12", "Need Wp 13", "Need Wp 14", "Need Wp 15", "Shield",
	"SP", "MP", "ReadyTime", "ResetTime", "AttackRange", "StateType",
	"AttrType", "Disable", "SuccessType", "SuccessValue", "TargetType", "ApplyRange",
	"MultiAttack", "KeepTime", "Weapon1", "Weapon2", "WeaponValue", "Bag",
	"Aggro", "DamageType", "DamageHP", "DamageSP", "DamageMP", "TimeDamageType",
	"TimeDamageHP", "TimeDamageSP", "TimeDamageMP", "AddDamageHP", "AddDamageSP", "AddDamageMP",
	"AT1", "AV1", "AT2", "AV2", "AT3", "AV3",
	"AT4", "AV4", "AT5", "AV5", "AT6", "AV6",
	"AT7", "AV7", "AT8", "AV8", "AT9", "AV9",
	"AT10", "AV10", "HealHP", "HealSP", "HealMP", "TimeHealHP",
	"TimeHealSP", "TimeHealMP", "DefenceType", "DefenceValue", "LimitHP", "FixRange",
	"ChangeType", "ChangeLevel",
}

var SkillColumnTags = []Tag{
	"", "", "", "", "", "", "", "",
	"figh", "defs", "rngr", "arch", "mage", "prst", "prvs", "rlvl",
	"grow", "pnts", "tpsh", "tpat", "tpef", "tpdt", "nwp1", "nwp2",
	"nwp3", "nwp4", "nwp5", "nwp6", "nwp7", "nwp8", "nwp9", "nwpA",
	"nwpB", "nwpC", "nwpD", "nwpE", "nwpF", "shld", "cssp", "csmp",
	"rdtm", "rstm", "atrg", "sttp", "attp", "disa", "sctp", "scvl",
	"trtp", "aprg", "mult", "kptm", "wpn1", "wpn2", "wval", "bags",
	"arrw", "dmtp", "dmhp", "dmsp", "dmmp", "tdtp", "tdhp", "tdsp",
	"tdmp", "adhp", "adsp", "admp", "abt1", "abv1", "abt2", "abv2",
	"abt3", "abv3", "abt4", "abv4", "abt5", "abv5", "abt6", "abv6",
	"abt7", "abv7", "abt8", "abv8", "abt9", "abv9", "abt:", "abv:",
	"hlhp", "hlsp", "hlmp", "thhp", "thsp", "thmp", "dftp", "dfvl",
	"lmhp", "fxrg", "chtp", "chlv",
}

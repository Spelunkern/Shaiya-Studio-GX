package sdata

// SkillTypeName names a skill type detail (getSkillTypeName).
func SkillTypeName(t int16) string {
	switch t {
	case 1062:
		return "Cure of Goddess"
	case 0:
		return "None"
	case 1:
		return "Weapon Power Up"
	case 2:
		return "Backpack"
	case 4:
		return "Passive Defense"
	case 5:
		return "Interpretation"
	case 6:
		return "Potential Defense"
	case 8:
		return "Weapon Mastery"
	case 10:
		return "Defense Power Up"
	case 21:
		return "Unique Hit Attack"
	case 22:
		return "Multiple Hit Attack"
	case 23:
		return "Eraser"
	case 25:
		return "HP, MP, SP Reduction"
	case 27:
		return "Energy Drain"
	case 41:
		return "Healing"
	case 42:
		return "Dispel"
	case 43:
		return "Sacrifice"
	case 61:
		return "Periodical Debuff"
	case 62:
		return "Subtracting Debuff"
	case 63:
		return "Flite"
	case 64:
		return "Stun"
	case 65:
		return "Immobilize"
	case 66:
		return "Sleep"
	case 67:
		return "Prevent Attack"
	case 68:
		return "Remove Attribute"
	case 69:
		return "Taunt"
	case 72:
		return "Provoke"
	case 74:
		return "Energy Blackhole"
	case 75:
		return "Mental Storm, Confusion"
	case 76:
		return "Soul Menace"
	case 77:
		return "Mental Storm, Distortion"
	case 78:
		return "unknow (mobs)"
	case 81:
		return "Periodical Heal"
	case 82:
		return "Buff, Noss"
	case 83:
		return "Damage Reflection"
	case 84:
		return "Fire thorn"
	case 85:
		return "Block Magic Attack"
	case 86:
		return "Block Shooting Attack"
	case 87:
		return "Health Assistant"
	case 88:
		return "Stealth"
	case 90:
		return "Evolution"
	case 91:
		return "Ability Exchange"
	case 92:
		return "Untouchable"
	case 93:
		return "Detection"
	case 94:
		return "Elemental Attack"
	case 95:
		return "Elemental Protection"
	case 96:
		return "Dance of Death"
	case 97:
		return "Mass Ambush"
	case 98:
		return "Etain Shield"
	case 101:
		return "Trap"
	case 102:
		return "Firebush"
	case 121:
		return "Resurrection"
	case 122:
		return "Soul Burn"
	case 123:
		return "Scouting"
	case 125:
		return "Town Portal"
	case 126:
		return "Mental Impact"
	}
	return "unknown"
}

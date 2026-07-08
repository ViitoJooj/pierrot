package compiler

// Operações de árvore usadas pelo render: localizar/substituir instâncias
// de componente e detectar tags capitalizadas que sobraram sem import.

// HasComponent informa se a árvore tem alguma instância <Name ... />
func HasComponent(nodes []Node, name string) bool {
	found := false
	walkNodes(nodes, func(n Node) {
		if ci, ok := n.(*ComponentInst); ok && ci.Name == name {
			found = true
		}
	})
	return found
}

// ReplaceComponents troca cada instância <Name /> pelos nós de expand(inst),
// em qualquer profundidade (dentro de elementos e blocos)
func ReplaceComponents(nodes []Node, name string, expand func(*ComponentInst) []Node) []Node {
	var out []Node
	for _, n := range nodes {
		switch v := n.(type) {
		case *ComponentInst:
			if v.Name == name {
				out = append(out, expand(v)...)
				continue
			}
			out = append(out, n)
		case *Element:
			el := *v
			el.Children = ReplaceComponents(v.Children, name, expand)
			out = append(out, &el)
		case *ForBlock:
			fb := *v
			fb.Children = ReplaceComponents(v.Children, name, expand)
			out = append(out, &fb)
		case *IfBlock:
			ib := *v
			ib.Then = ReplaceComponents(v.Then, name, expand)
			ib.Else = ReplaceComponents(v.Else, name, expand)
			out = append(out, &ib)
		default:
			out = append(out, n)
		}
	}
	return out
}

// UnknownTags coleta, dedup na ordem, os nomes capitalizados que sobraram
// na árvore: instâncias não expandidas, elementos com tag maiúscula e Slot
// fora do layout (mesma detecção do unknownTagRe antigo)
func UnknownTags(nodes []Node) []string {
	var out []string
	seen := map[string]bool{}
	add := func(name string) {
		if name != "" && name[0] >= 'A' && name[0] <= 'Z' && !seen[name] {
			seen[name] = true
			out = append(out, name)
		}
	}
	walkNodes(nodes, func(n Node) {
		switch v := n.(type) {
		case *ComponentInst:
			add(v.Name)
		case *Element:
			add(v.Tag)
		case *Slot:
			add("Slot")
		}
	})
	return out
}

// SpliceSlots troca cada <Slot /> da árvore pelos nós da página
func SpliceSlots(nodes []Node, page []Node) []Node {
	var out []Node
	for _, n := range nodes {
		switch v := n.(type) {
		case *Slot:
			out = append(out, page...)
		case *Element:
			el := *v
			el.Children = SpliceSlots(v.Children, page)
			out = append(out, &el)
		case *ForBlock:
			fb := *v
			fb.Children = SpliceSlots(v.Children, page)
			out = append(out, &fb)
		case *IfBlock:
			ib := *v
			ib.Then = SpliceSlots(v.Then, page)
			ib.Else = SpliceSlots(v.Else, page)
			out = append(out, &ib)
		default:
			out = append(out, n)
		}
	}
	return out
}

// walkNodes visita todos os nós da árvore em profundidade
func walkNodes(nodes []Node, fn func(Node)) {
	for _, n := range nodes {
		fn(n)
		switch v := n.(type) {
		case *Element:
			walkNodes(v.Children, fn)
		case *ForBlock:
			walkNodes(v.Children, fn)
		case *IfBlock:
			walkNodes(v.Then, fn)
			walkNodes(v.Else, fn)
		}
	}
}

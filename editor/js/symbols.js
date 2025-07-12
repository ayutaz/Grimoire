// Symbol definitions and management
class SymbolManager {
    constructor() {
        this.symbols = {
            // Structure Elements
            'outer-circle': { symbol: '⭕', name: 'Outer Circle', category: 'structure', defaultSize: 200 },
            'inner-circle': { symbol: '○', name: 'Inner Circle', category: 'structure', defaultSize: 100 },
            'double-circle': { symbol: '◎', name: 'Double Circle', category: 'structure', defaultSize: 200 },
            'pentagram': { symbol: '⭐', name: 'Pentagram', category: 'structure', defaultSize: 80 },
            'hexagram': { symbol: '✡', name: 'Hexagram', category: 'structure', defaultSize: 80 },
            'octagram': { symbol: '✦', name: 'Octagram', category: 'structure', defaultSize: 80 },
            'triangle': { symbol: '△', name: 'Triangle', category: 'structure', defaultSize: 60 },
            'square': { symbol: '□', name: 'Square', category: 'structure', defaultSize: 60 },
            
            // Mystical Operators
            'fusion': { symbol: '⟐', name: 'Fusion (Add)', category: 'operator', defaultSize: 50 },
            'separation': { symbol: '⟑', name: 'Separation (Subtract)', category: 'operator', defaultSize: 50 },
            'amplify': { symbol: '✦', name: 'Amplify (Multiply)', category: 'operator', defaultSize: 50 },
            'divide': { symbol: '⟠', name: 'Division', category: 'operator', defaultSize: 50 },
            'transfer': { symbol: '⟷', name: 'Transfer (Assign)', category: 'operator', defaultSize: 50 },
            'seal': { symbol: '⊗', name: 'Seal (Constant)', category: 'operator', defaultSize: 50 },
            'cycle': { symbol: '⟳', name: 'Cycle (Loop)', category: 'operator', defaultSize: 50 },
            
            // Comparison Symbols
            'equal': { symbol: '=', name: 'Equal', category: 'comparison', defaultSize: 40 },
            'not-equal': { symbol: '≠', name: 'Not Equal', category: 'comparison', defaultSize: 40 },
            'less': { symbol: '<', name: 'Less Than', category: 'comparison', defaultSize: 40 },
            'greater': { symbol: '>', name: 'Greater Than', category: 'comparison', defaultSize: 40 },
            'less-equal': { symbol: '≤', name: 'Less or Equal', category: 'comparison', defaultSize: 40 },
            'greater-equal': { symbol: '≥', name: 'Greater or Equal', category: 'comparison', defaultSize: 40 },
            
            // Logic Symbols
            'and': { symbol: '⊕', name: 'AND', category: 'logic', defaultSize: 50 },
            'or': { symbol: '⊖', name: 'OR', category: 'logic', defaultSize: 50 },
            'not': { symbol: '⊗', name: 'NOT', category: 'logic', defaultSize: 50 },
            'xor': { symbol: '⊙', name: 'XOR', category: 'logic', defaultSize: 50 },
            
            // Energy Nodes
            'hex-crystal': { symbol: '⬢', name: 'Branch Point', category: 'node', defaultSize: 60 },
            'square-crystal': { symbol: '◈', name: 'Aggregation', category: 'node', defaultSize: 60 },
            'penta-crystal': { symbol: '⬟', name: 'Transform', category: 'node', defaultSize: 60 },
            'star-crystal': { symbol: '✧', name: 'Amplify', category: 'node', defaultSize: 60 },
            
            // Special Symbols
            'sun': { symbol: '☀', name: 'Start/True', category: 'special', defaultSize: 70 },
            'moon': { symbol: '☾', name: 'False/Alt', category: 'special', defaultSize: 70 },
            'star': { symbol: '☆', name: 'Output', category: 'special', defaultSize: 60 },
            'double-node': { symbol: '○○', name: 'Function', category: 'special', defaultSize: 80 },
            'note': { symbol: '♪', name: 'Sound', category: 'special', defaultSize: 50 },
            'envelope': { symbol: '✉', name: 'Message', category: 'special', defaultSize: 50 },
            'check': { symbol: '✓', name: 'Success', category: 'special', defaultSize: 50 },
            'plus': { symbol: '+', name: 'Plus', category: 'special', defaultSize: 40 },
            'data-node': { symbol: '□•', name: 'Data Node', category: 'special', defaultSize: 60 }
        };
    }
    
    getSymbol(type) {
        return this.symbols[type] || null;
    }
    
    getSymbolsByCategory(category) {
        return Object.entries(this.symbols)
            .filter(([key, value]) => value.category === category)
            .map(([key, value]) => ({ type: key, ...value }));
    }
    
    isCircleType(type) {
        return ['outer-circle', 'inner-circle', 'double-circle'].includes(type);
    }
    
    isConnectionPoint(type) {
        // These symbols can have connections
        return !['outer-circle'].includes(type);
    }
}

// Create global instance
window.symbolManager = new SymbolManager();
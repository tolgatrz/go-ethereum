contract HashVote {
    struct Period {
        mapping(bytes32 => uint) entries;
        bytes32[] indices;
    }
    
    Period[] periods;
    
    modifier mustBeVoter() {
        if( !canVote[msg.sender] ) throw;
    }
    
    uint256 public start;
    function HashVote() {
        canVote[msg.sender] = true;
        
        start = block.number;
        
        Period p = periods[periods.length++];
        
        bytes32 hash = block.blockhash(block.number-1);
        p.entries[hash]++;
        p.indices.push(hash);
    }
    
    function getIndex() internal returns(uint) {
        return block.number - start;
    }
    
    function vote(bytes32 hash) {
        uint index = getIndex();
        if( periods.length <= index ) periods.length++;
        
        Period period = periods[index];
        
        if(period.entries[hash] == 0) period.indices.push(hash);
        period.entries[hash]++;
    }
    
    function getCanonHash() constant returns(bytes32) {
        Period period = periods[periods.length-1];
        
        bytes32 best;
        for(uint i = 0; i < period.indices.length; i++) {
            if(period.entries[best] < period.entries[period.indices[i]]) {
                best = period.indices[i];
            }
        }
        return best;
    }

    // returns the amount of votes within this period
    function getVotePeriod(period uint) returns(uint) {
	    return periods[period].indices.length;
    }

    mapping(address => bool) public canVote;    
    function addVoter(address addr) {
        if( !canVote[msg.sender] ) return;
        
        canVote[addr] = true;
    }
    
    function getSize() constant returns(uint) {
        return periods.length;
    }
    
    function getEntry(uint p, uint n) constant returns(bytes32) {
        Period period = periods[p];
        return period.indices[n];
    }
}

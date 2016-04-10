contract Consensus {
	struct Period {
		mapping(bytes32 => uint) entries;
		bytes32[] indices;
	}

	Period[] periods;

	modifier mustBeVoter() {
		if( !canVote[msg.sender] ) throw;
	}

	uint public start;
	/*function Consensus() {
		canVote[msg.sender] = true;

		start = block.number;

		Period p = periods[periods.length++];

		bytes32 hash = block.blockhash(block.number-1);
		p.entries[hash]++;
		p.indices.push(hash);
	}*/

	function vote(bytes32 hash) {
		if( periods.length < block.number ) periods.length++;
		Period period = periods[block.number-1];
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

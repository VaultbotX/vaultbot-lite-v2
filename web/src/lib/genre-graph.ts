import { CommunityPartition } from "./louvain";
import { assignCommunityColors, COMMUNITY_PALETTE, edgeWidth, nodeSize } from "./graph";

export class Community {
	readonly id: number;
	readonly color: string;
	readonly memberIds: ReadonlySet<number>;

	constructor(id: number, color: string, memberIds: ReadonlySet<number>) {
		this.id = id;
		this.color = color;
		this.memberIds = memberIds;
	}

	get memberCount(): number { return this.memberIds.size; }
}

export class GenreNode {
	readonly genreId: number;
	readonly name: string;
	readonly artistCount: number;
	readonly community: Community;

	constructor(genreId: number, name: string, artistCount: number, community: Community) {
		this.genreId = genreId;
		this.name = name;
		this.artistCount = artistCount;
		this.community = community;
	}

	displaySize(maxArtistCount: number): number {
		return nodeSize(this.artistCount, maxArtistCount);
	}

	get displayColor(): string { return this.community.color; }
}

export class GenreEdge {
	readonly source: GenreNode;
	readonly target: GenreNode;
	readonly sharedArtistCount: number;

	constructor(source: GenreNode, target: GenreNode, sharedArtistCount: number) {
		this.source = source;
		this.target = target;
		this.sharedArtistCount = sharedArtistCount;
	}

	displayWidth(maxShared: number): number {
		return edgeWidth(this.sharedArtistCount, maxShared);
	}

	displayOpacity(maxShared: number): number {
		return 0.15 + 0.5 * Math.sqrt(this.sharedArtistCount / maxShared);
	}
}

export interface NodeDisplay {
	id: string;
	label: string;
	size: number;
	color: string;
	genreId: number;
}

export interface EdgeDisplay {
	sourceId: string;
	targetId: string;
	width: number;
	opacity: number;
	shared: number;
}

export class GenreGraph {
	readonly nodes: readonly GenreNode[];
	readonly edges: readonly GenreEdge[];
	readonly communities: ReadonlyMap<number, Community>;
	private readonly _maxArtistCount: number;
	private readonly _maxShared: number;

	private constructor(
		nodes: GenreNode[],
		edges: GenreEdge[],
		communities: Map<number, Community>,
	) {
		this.nodes = nodes;
		this.edges = edges;
		this.communities = communities;
		this._maxArtistCount = Math.max(...nodes.map((n) => n.artistCount), 1);
		this._maxShared = Math.max(...edges.map((e) => e.sharedArtistCount), 1);
	}

	nodeDisplays(): NodeDisplay[] {
		return this.nodes.map((node) => ({
			id: String(node.genreId),
			label: node.name,
			size: node.displaySize(this._maxArtistCount),
			color: node.displayColor,
			genreId: node.genreId,
		}));
	}

	edgeDisplays(): EdgeDisplay[] {
		return this.edges.map((edge) => ({
			sourceId: String(edge.source.genreId),
			targetId: String(edge.target.genreId),
			width: edge.displayWidth(this._maxShared),
			opacity: edge.displayOpacity(this._maxShared),
			shared: edge.sharedArtistCount,
		}));
	}

	static build(
		vertices: Array<{ genre_id: number; name: string; artist_count: number }>,
		apiEdges: Array<{
			source_genre_id: number;
			target_genre_id: number;
			shared_artist_count: number;
		}>,
		partition: CommunityPartition,
		palette: readonly string[] = COMMUNITY_PALETTE,
	): GenreGraph {
		const colorMap = assignCommunityColors(partition.communityIds, palette);

		const communityMap = new Map<number, Community>();
		for (const commId of partition.communityIds) {
			communityMap.set(
				commId,
				new Community(commId, colorMap.get(commId) ?? palette[0], partition.membersOf(commId)),
			);
		}

		const nodeMap = new Map<number, GenreNode>();
		for (const vertex of vertices) {
			const commId = partition.communityOf(vertex.genre_id);
			const community = commId !== undefined ? communityMap.get(commId) : undefined;
			if (!community) continue;
			nodeMap.set(
				vertex.genre_id,
				new GenreNode(vertex.genre_id, vertex.name, vertex.artist_count, community),
			);
		}

		const edges: GenreEdge[] = [];
		for (const apiEdge of apiEdges) {
			const source = nodeMap.get(apiEdge.source_genre_id);
			const target = nodeMap.get(apiEdge.target_genre_id);
			if (source && target) {
				edges.push(new GenreEdge(source, target, apiEdge.shared_artist_count));
			}
		}

		return new GenreGraph([...nodeMap.values()], edges, communityMap);
	}
}

package models

// GenreMusic represents Shazam chart music genres.
type GenreMusic string

const (
	GenrePop              GenreMusic = "pop"
	GenreHipHopRap        GenreMusic = "hip-hop-rap"
	GenreDance            GenreMusic = "dance"
	GenreElectronic       GenreMusic = "electronic"
	GenreRnBSoul          GenreMusic = "randb-soul"
	GenreAlternative      GenreMusic = "alternative"
	GenreRock             GenreMusic = "rock"
	GenreLatin            GenreMusic = "latin"
	GenreFilmTVStage      GenreMusic = "film-tv-and-stage"
	GenreCountry          GenreMusic = "country"
	GenreAfroBeats        GenreMusic = "afrobeats"
	GenreWorldwide        GenreMusic = "worldwide"
	GenreReggaeDanceHall  GenreMusic = "reggae-dancehall"
	GenreHouse            GenreMusic = "house"
	GenreKPop             GenreMusic = "k-pop"
	GenreFrenchPop        GenreMusic = "french-pop"
	GenreSingerSongwriter GenreMusic = "singer-songwriter"
	GenreRegionalMexicano GenreMusic = "regional-mexicano"
)

// ArtistExtend represents extension fields when querying artist profile.
type ArtistExtend string

const (
	ArtistExtendBio              ArtistExtend = "artistBio"
	ArtistExtendBornOrFormed     ArtistExtend = "bornOrFormed"
	ArtistExtendEditorialArtwork ArtistExtend = "editorialArtwork"
	ArtistExtendOrigin           ArtistExtend = "origin"
)

// ArtistView represents profile views when querying artist profile.
type ArtistView string

const (
	ArtistViewFullAlbums     ArtistView = "full-albums"
	ArtistViewFeaturedAlbums ArtistView = "featured-albums"
	ArtistViewLatestRelease  ArtistView = "latest-release"
	ArtistViewTopMusicVideos ArtistView = "top-music-videos"
	ArtistViewSimilarArtists ArtistView = "similar-artists"
	ArtistViewTopSongs       ArtistView = "top-songs"
	ArtistViewPlaylists      ArtistView = "playlists"
)

// Device represents device platform identifier sent to Shazam.
type Device string

const (
	DeviceIPhone  Device = "iphone"
	DeviceAndroid Device = "android"
	DeviceWeb     Device = "web"
)

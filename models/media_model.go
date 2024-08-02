package models

import (
	"encoding/json"
)

func UnmarshalMediaInfo(data []byte) (MediaInfo, error) {
	var r MediaInfo
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *MediaInfo) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type MediaInfo struct {
	CreatingLibrary *CreatingLibrary `json:"creatingLibrary,omitempty"`
	Media           *Media           `json:"media,omitempty"`
}

type CreatingLibrary struct {
	Name    *string `json:"name,omitempty"`
	Version *string `json:"version,omitempty"`
	URL     *string `json:"url,omitempty"`
}

type Media struct {
	Ref   *string `json:"@ref,omitempty"`
	Track []Track `json:"track,omitempty"`
}

type Track struct {
	Type                           *string `json:"@type,omitempty"`
	VideoCount                     *string `json:"VideoCount,omitempty"`
	AudioCount                     *string `json:"AudioCount,omitempty"`
	OtherCount                     *string `json:"OtherCount,omitempty"`
	FileExtension                  *string `json:"FileExtension,omitempty"`
	Format                         *string `json:"Format,omitempty"`
	FormatProfile                  *string `json:"Format_Profile,omitempty"`
	CodecID                        *string `json:"CodecID,omitempty"`
	CodecIDVersion                 *string `json:"CodecID_Version,omitempty"`
	CodecIDCompatible              *string `json:"CodecID_Compatible,omitempty"`
	FileSize                       *string `json:"FileSize,omitempty"`
	Duration                       *string `json:"Duration,omitempty"`
	OverallBitRate                 *string `json:"OverallBitRate,omitempty"`
	FrameRate                      *string `json:"FrameRate,omitempty"`
	FrameCount                     *string `json:"FrameCount,omitempty"`
	StreamSize                     *string `json:"StreamSize,omitempty"`
	HeaderSize                     *string `json:"HeaderSize,omitempty"`
	DataSize                       *string `json:"DataSize,omitempty"`
	FooterSize                     *string `json:"FooterSize,omitempty"`
	IsStreamable                   *string `json:"IsStreamable,omitempty"`
	EncodedDate                    *string `json:"Encoded_Date,omitempty"`
	TaggedDate                     *string `json:"Tagged_Date,omitempty"`
	FileModifiedDate               *string `json:"File_Modified_Date,omitempty"`
	FileModifiedDateLocal          *string `json:"File_Modified_Date_Local,omitempty"`
	EncodedLibrary                 *string `json:"Encoded_Library,omitempty"`
	EncodedLibraryName             *string `json:"Encoded_Library_Name,omitempty"`
	Extra                          *Extra  `json:"extra,omitempty"`
	StreamOrder                    *string `json:"StreamOrder,omitempty"`
	ID                             *string `json:"ID,omitempty"`
	FormatLevel                    *string `json:"Format_Level,omitempty"`
	FormatSettingsCABAC            *string `json:"Format_Settings_CABAC,omitempty"`
	FormatSettingsRefFrames        *string `json:"Format_Settings_RefFrames,omitempty"`
	SourceDuration                 *string `json:"Source_Duration,omitempty"`
	SourceDurationLastFrame        *string `json:"Source_Duration_LastFrame,omitempty"`
	BitRate                        *string `json:"BitRate,omitempty"`
	Width                          *string `json:"Width,omitempty"`
	Height                         *string `json:"Height,omitempty"`
	SampledWidth                   *string `json:"Sampled_Width,omitempty"`
	SampledHeight                  *string `json:"Sampled_Height,omitempty"`
	PixelAspectRatio               *string `json:"PixelAspectRatio,omitempty"`
	DisplayAspectRatio             *string `json:"DisplayAspectRatio,omitempty"`
	Rotation                       *string `json:"Rotation,omitempty"`
	FrameRateMode                  *string `json:"FrameRate_Mode,omitempty"`
	FrameRateNum                   *string `json:"FrameRate_Num,omitempty"`
	FrameRateDen                   *string `json:"FrameRate_Den,omitempty"`
	ColorSpace                     *string `json:"ColorSpace,omitempty"`
	ChromaSubsampling              *string `json:"ChromaSubsampling,omitempty"`
	BitDepth                       *string `json:"BitDepth,omitempty"`
	ScanType                       *string `json:"ScanType,omitempty"`
	SourceStreamSize               *string `json:"Source_StreamSize,omitempty"`
	Title                          *string `json:"Title,omitempty"`
	ColourDescriptionPresent       *string `json:"colour_description_present,omitempty"`
	ColourDescriptionPresentSource *string `json:"colour_description_present_Source,omitempty"`
	ColourRange                    *string `json:"colour_range,omitempty"`
	ColourRangeSource              *string `json:"colour_range_Source,omitempty"`
	ColourPrimaries                *string `json:"colour_primaries,omitempty"`
	ColourPrimariesSource          *string `json:"colour_primaries_Source,omitempty"`
	TransferCharacteristics        *string `json:"transfer_characteristics,omitempty"`
	TransferCharacteristicsSource  *string `json:"transfer_characteristics_Source,omitempty"`
	MatrixCoefficients             *string `json:"matrix_coefficients,omitempty"`
	MatrixCoefficientsSource       *string `json:"matrix_coefficients_Source,omitempty"`
	FormatAdditionalFeatures       *string `json:"Format_AdditionalFeatures,omitempty"`
	BitRateMode                    *string `json:"BitRate_Mode,omitempty"`
	Channels                       *string `json:"Channels,omitempty"`
	ChannelPositions               *string `json:"ChannelPositions,omitempty"`
	ChannelLayout                  *string `json:"ChannelLayout,omitempty"`
	SamplesPerFrame                *string `json:"SamplesPerFrame,omitempty"`
	SamplingRate                   *string `json:"SamplingRate,omitempty"`
	SamplingCount                  *string `json:"SamplingCount,omitempty"`
	SourceFrameCount               *string `json:"Source_FrameCount,omitempty"`
	CompressionMode                *string `json:"Compression_Mode,omitempty"`
	Typeorder                      *string `json:"@typeorder,omitempty"`
	TrackType                      *string `json:"Type,omitempty"`
}

type Extra struct {
	COMAppleQuicktimeCreationdate      *string `json:"com_apple_quicktime_creationdate,omitempty"`
	COMApplePhotosOriginatingSignature *string `json:"com_apple_photos_originating_signature,omitempty"`
	COMAppleQuicktimeLocationISO6709   *string `json:"com_apple_quicktime_location_ISO6709,omitempty"`
	Metas                              *string `json:"Metas,omitempty"`
	SourceDelay                        *string `json:"Source_Delay,omitempty"`
	SourceDelaySource                  *string `json:"Source_Delay_Source,omitempty"`
	CodecConfigurationBox              *string `json:"CodecConfigurationBox,omitempty"`
	EncodedDate                        *string `json:"Encoded_Date,omitempty"`
	TaggedDate                         *string `json:"Tagged_Date,omitempty"`
}

package service

import (
	"context"
	"fmt"

	"google.golang.org/api/youtube/v3"
)

type YouTubeService struct {
	googleOAuth *GoogleOAuthService
}

func NewYouTubeService(googleOAuth *GoogleOAuthService) *YouTubeService {
	return &YouTubeService{
		googleOAuth: googleOAuth,
	}
}

// SearchVideos searches for videos on YouTube.
func (s *YouTubeService) SearchVideos(ctx context.Context, query string, channelID string, limit int64) (*youtube.SearchListResponse, error) {
	youtubeService, err := s.googleOAuth.GetYouTubeService(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get YouTube service: %w", err)
	}

	call := youtubeService.Search.List([]string{"id", "snippet"}).Q(query).Type("video").MaxResults(limit)
	if channelID != "" {
		call.ChannelId(channelID)
	}

	response, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to search videos: %w", err)
	}

	return response, nil
}

// GetVideoMetadata retrieves detailed information about a specific video.
func (s *YouTubeService) GetVideoMetadata(ctx context.Context, videoID string) (*youtube.VideoListResponse, error) {
	youtubeService, err := s.googleOAuth.GetYouTubeService(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get YouTube service: %w", err)
	}

	call := youtubeService.Videos.List([]string{"snippet", "statistics", "contentDetails"}).Id(videoID)

	response, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get video metadata: %w", err)
	}

	return response, nil
}

// GetVideoComments retrieves top-level comment threads for a video.
func (s *YouTubeService) GetVideoComments(ctx context.Context, videoID string, sortBy string, limit int64) (*youtube.CommentThreadListResponse, error) {
	youtubeService, err := s.googleOAuth.GetYouTubeService(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get YouTube service: %w", err)
	}

	call := youtubeService.CommentThreads.List([]string{"snippet", "replies"}).VideoId(videoID).Order(sortBy).MaxResults(limit)

	response, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get video comments: %w", err)
	}

	return response, nil
}

// ReplyToComment posts a reply to a specific comment.
// This is an owner-only action.
func (s *YouTubeService) ReplyToComment(ctx context.Context, parentID string, text string) (*youtube.Comment, error) {
	youtubeService, err := s.googleOAuth.GetYouTubeService(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get YouTube service: %w", err)
	}

	comment := &youtube.Comment{
		Snippet: &youtube.CommentSnippet{
			ParentId:     parentID,
			TextOriginal: text,
		},
	}

	call := youtubeService.Comments.Insert([]string{"snippet"}, comment)
	response, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to post comment: %w", err)
	}

	return response, nil
}

// AddVideoToPlaylist adds a video to a specific playlist.
// This is an owner-only action.
func (s *YouTubeService) AddVideoToPlaylist(ctx context.Context, playlistID string, videoID string) (*youtube.PlaylistItem, error) {
	youtubeService, err := s.googleOAuth.GetYouTubeService(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get YouTube service: %w", err)
	}

	playlistItem := &youtube.PlaylistItem{
		Snippet: &youtube.PlaylistItemSnippet{
			PlaylistId: playlistID,
			ResourceId: &youtube.ResourceId{
				Kind:    "youtube#video",
				VideoId: videoID,
			},
		},
	}

	call := youtubeService.PlaylistItems.Insert([]string{"snippet"}, playlistItem)
	response, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to add video to playlist: %w", err)
	}

	return response, nil
}

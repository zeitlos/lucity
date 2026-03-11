import gql from 'graphql-tag';

export const SearchImagesQuery = gql`
  query SearchImages($query: String!) {
    searchImages(query: $query) {
      name
      description
      starCount
      pullCount
      official
    }
  }
`;

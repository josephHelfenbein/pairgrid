<!-- Improved compatibility of back to top link: See: https://github.com/othneildrew/Best-README-Template/pull/73 -->
<a id="readme-top"></a>
<!--
*** Thanks for checking out the Best-README-Template. If you have a suggestion
*** that would make this better, please fork the repo and create a pull request
*** or simply open an issue with the tag "enhancement".
*** Don't forget to give the project a star!
*** Thanks again! Now go create something AMAZING! :D
-->



<!-- PROJECT SHIELDS -->
<!--
*** I'm using markdown "reference style" links for readability.
*** Reference links are enclosed in brackets [ ] instead of parentheses ( ).
*** See the bottom of this document for the declaration of the reference variables
*** for contributors-url, forks-url, etc. This is an optional, concise syntax you may use.
*** https://www.markdownguide.org/basic-syntax/#reference-style-links
-->
[![Contributors][contributors-shield]][contributors-url]
[![Stargazers][stars-shield]][stars-url]
[![Issues][issues-shield]][issues-url]
[![Apache License][license-shield]][license-url]
[![LinkedIn][linkedin-shield]][linkedin-url-joseph]



<!-- PROJECT LOGO -->
<br />
<div align="center">
  <a href="https://github.com/josephHelfenbein/pairgrid">
    <img src="/public/pairgrid-icon.svg" alt="Logo" width="80" height="80">
  </a>

<h3 align="center">PairGrid</h3>

  <p align="center">
    PairGrid is a platform for connecting developers through smart matchmaking, real-time chat, and video collaboration, designed to facilitate coding partnerships and project collaboration.
    <br />
    <br />
    <a href="https://www.pairgrid.com">Visit</a>
    ·
    <a href="https://github.com/josephHelfenbein/pairgrid/issues/new?labels=bug&template=bug-report---.md">Report Bug</a>
    ·
    <a href="https://github.com/josephHelfenbein/pairgrid/issues/new?labels=enhancement&template=feature-request---.md">Request Feature</a>
  </p>
</div>



<!-- TABLE OF CONTENTS -->
<details>
  <summary>Table of Contents</summary>
  <ol>
    <li>
      <a href="#about-the-project">About The Project</a>
      <ul>
        <li><a href="#built-with">Built With</a></li>
      </ul>
    </li>
    <li>
      <a href="#getting-started">Getting Started</a>
      <ul>
        <li><a href="#prerequisites">Prerequisites</a></li>
        <li><a href="#installation">Installation</a></li>
      </ul>
    </li>
    <li><a href="#license">License</a></li>
  </ol>
</details>



<!-- ABOUT THE PROJECT -->
## About the Project

(This project is still under construction)

PairGrid is a real-time matchmaking platform designed to connect developers with compatible coding partners. Whether you're looking for collaborators who share your interests, tech stack, or coding goals, PairGrid's smart matchmaking system has you covered. Features include real-time chat, video calls with screen sharing, and seamless collaboration tools to help you build amazing projects together. Currently under construction, PairGrid aims to empower developers to connect, code, and create like never before.

### How does it work?

Users are able to sign in using Clerk authentication. They then are able to change their preferences and edit their bio, interests, known programming languages, specialty, and occupation. Through this, users are shown on the networking tab by similarity to the user. The user can send friend requests to other users, and see friend requests to themselves on the chat tab. They can use real-time chat messages, video chat, or screensharing with friends.

The Clerk authentication system connects to the Hasura database using webhooks on user creation, modification, or deletion. Using Go serverless endpoints and GraphQL, the database is able to be updated and queried.
For real-time chat, the frontend encrypts the messages using a server-side key and sends it to Hasura using a Go serverless endpoint and GraphQL. In the same endpoint, it sends the message information to Pusher. If another user is listening to the Pusher channel (has the chat with the user that sent the message open), then it decrypts the message and displays it. When a chat is opened, it also gets all messages in the conversation from the Hasura database.

For ranking the users by similarity, it uses a SQL function to sort the users from Hasura by weights in descending order, with the weights computed by adding similarities in languages, interests, occupation and specialty. Video chat and screensharing are still to be implemented.





### Built With

* [![Nuxt.js][Nuxt.js]][Nuxt.js-url]
* [![Vue.js][Vue.js]][Vue.js-url]
* [![Go][Go]][Go-url]
* [![Tailwind][Tailwind]][Tailwind-url]
* [![Shadcn][Shadcn]][Shadcn-url]
* [![Clerk][Clerk]][Clerk-url]
* [![Hasura][Hasura]][Hasura-url]
* [![Graphql][Graphql]][Graphql-url]
* [![Pusher][Pusher]][Pusher-url]



<p align="right">(<a href="#readme-top">back to top</a>)</p>




<!-- GETTING STARTED -->
## Getting Started

Here are the steps to run the project locally if you want to develop your own project.

### Prerequisites

* pnpm
  ```sh
  pnpm self-update
  ```


### Installation

1. Fork the repository and set it up as a project on Vercel or another hosting platform

2. Install packages
   ```sh
   pnpm install
   ```

3. Create a Hasura account at [https://hasura.io/](https://hasura.io/) and start a project on the legacy Hasura dashboard. Get the API keys `HASURA_GRAPHQL_URL, HASURA_GRAPHQL_ADMIN_SECRET` and put them in the environment variables. Additionally, create tables "users", "friends", and "messages" with the same columns found in the [Go serverless endpoints](https://github.com/josephHelfenbein/pairgrid/tree/main/api). Create an empty table called 'similarity_result' with the columns:
    ```
    id- text, primary key, unique
    name- text
    email- text
    bio- text, nullable
    language- text[], nullable
    specialty- text, nullable
    interests- text[], nullable
    occupation- text, nullable
    profile_picture- text
    similarity_score- bigint
    ```
    Run this in the SQL tab to create an SQL function for getting the similarity between two users:
    ```sql
    CREATE OR REPLACE FUNCTION calculate_similarity_score(user_id text)
    RETURNS SETOF similarity_result AS $$
    WITH target_user AS (
        SELECT
            specialty,
            occupation,
            language,
            interests
        FROM users
        WHERE id = user_id
    )
    SELECT
        u.id,
        u.name,
        u.email,
        u.bio,
        u.language,
        u.specialty,
        u.interests,
        u.occupation,
        u.profile_picture,
        (
            SELECT COUNT(*) 
            FROM UNNEST(u.interests) AS user_interest
            INNER JOIN UNNEST((SELECT interests FROM target_user)) AS target_interest
            ON user_interest = target_interest
        ) +
        (
            SELECT COUNT(*) 
            FROM UNNEST(u.language) AS user_language
            INNER JOIN UNNEST((SELECT language FROM target_user)) AS target_language
            ON user_language = target_language
        ) +
        CASE 
            WHEN u.specialty = (SELECT specialty FROM target_user) THEN 4
            ELSE 0
        END +
        CASE 
            WHEN u.occupation = (SELECT occupation FROM target_user) THEN 2
            ELSE 0
        END AS similarity_score
    FROM users u
    WHERE u.id != user_id
    ORDER BY similarity_score DESC;
    $$ LANGUAGE sql;
    ```
  
4. Create a Pusher account at [https://pusher.com/](https://pusher.com/) and start a project. Get the API keys `PUSHER_APP_ID, PUSHER_APP_KEY, PUSHER_APP_SECRET` and put them in the environment variables. 

5. Create a Clerk account at [https://clerk.com/](https://clerk.com/), and create a project. Get the API keys `NUXT_PUBLIC_CLERK_PUBLISHABLE_KEY, NUXT_CLERK_SECRET_KEY`
    and put them in the environment variables. Additionally, create webhooks on Clerk, one of them with endpoint {yourdomain}/api/userdelete/userdelete with a subscribed event of user.deleted, and one of them with endpoint {yourdomain}/api/userupdate with subscribed events user.created and user.updated. Get the signing secrets for both and set them to environment variables `UPDATE_SIGNING_SECRET, DELETE_SIGNING_SECRET`.

 6. Create a random server-side encryption key using OpenSSL
    ```bash
    openssl rand -hex 32
    ```
    and store it in the environment variables under `ENCRYPTION_KEY`

7. You can run the website locally with
    ```sh
    npm run dev
    ```
    or, if hosting on Vercel, with
    ```sh
    vercel dev
    ```








<!-- LICENSE -->
## License

Distributed under the Apache 2.0 License. See `LICENSE.txt` for more information.


* [Best README Template](https://github.com/othneildrew/Best-README-Template)

<p align="right">(<a href="#readme-top">back to top</a>)</p>





<!-- MARKDOWN LINKS & IMAGES -->
<!-- https://www.markdownguide.org/basic-syntax/#reference-style-links -->
[contributors-shield]: https://img.shields.io/github/contributors/josephHelfenbein/pairgrid.svg?style=for-the-badge
[contributors-url]: https://github.com/josephHelfenbein/pairgrid/graphs/contributors
[forks-shield]: https://img.shields.io/github/forks/josephHelfenbein/pairgrid.svg?style=for-the-badge
[forks-url]: https://github.com/josephHelfenbein/pairgrid/network/members
[stars-shield]: https://img.shields.io/github/stars/josephHelfenbein/pairgrid.svg?style=for-the-badge
[stars-url]: https://github.com/josephHelfenbein/pairgrid/stargazers
[issues-shield]: https://img.shields.io/github/issues/josephHelfenbein/pairgrid.svg?style=for-the-badge
[issues-url]: https://github.com/josephHelfenbein/pairgrid/issues
[license-shield]: https://img.shields.io/github/license/josephHelfenbein/pairgrid.svg?style=for-the-badge
[license-url]: https://github.com/josephHelfenbein/pairgrid/blob/master/LICENSE.txt
[linkedin-shield]: https://img.shields.io/badge/LinkedIn-0A66C2.svg?style=for-the-badge&logo=linkedin&logoColor=white
[linkedin-url-joseph]: https://linkedin.com/in/joseph-j-helfenbein
[product-screenshot]: images/screenshot.png
[Nuxt.js]: https://img.shields.io/badge/Nuxt.js-00DC82?style=for-the-badge&logo=nuxt&logoColor=white
[Nuxt.js-url]: https://nuxt.com/
[Vue.js]: https://img.shields.io/badge/Vue.js-4FC08D?style=for-the-badge&logo=vue.js&logoColor=white
[Vue.js-url]: https://vuejs.org/
[Go]: https://img.shields.io/badge/go-00ADD8?style=for-the-badge&logo=go&logoColor=white
[Go-url]: https://go.dev/
[Tailwind]: https://img.shields.io/badge/Tailwind%20CSS-06B6D4?style=for-the-badge&logo=tailwindcss&logoColor=white
[Tailwind-url]: https://tailwindcss.com/
[Shadcn]: https://img.shields.io/badge/shadcn%E2%80%93vue-000000?style=for-the-badge&logo=shadcn/ui&logoColor=4FC08D
[Shadcn-url]: https://www.shadcn-vue.com/
[Clerk]: https://img.shields.io/badge/clerk-6C47FF?logo=clerk&style=for-the-badge&logoColor=white
[Clerk-url]: https://clerk.com/
[Hasura]: https://img.shields.io/badge/hasura-1EB4D4?logo=hasura&style=for-the-badge&logoColor=white
[Hasura-url]: https://hasura.io/
[Graphql]: https://img.shields.io/badge/graphql-E10098?style=for-the-badge&logo=graphql&logoColor=white
[Graphql-url]: https://graphql.org/
[Pusher]: https://img.shields.io/badge/pusher-300D4F?style=for-the-badge&logo=pusher&logoColor=white
[Pusher-url]: https://pusher.com/

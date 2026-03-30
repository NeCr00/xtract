// ============================================
// Layer 3: Framework-Aware Extraction Test Data
// ============================================

// Technique 34: React Router
<Route path="/dashboard" component={Dashboard} />;
<Route exact path="/users/:id" component={UserProfile} />;
<Route path="/settings/notifications" />;
<Link to="/about">About</Link>;
<NavLink to="/contact">Contact</NavLink>;
navigate('/checkout/payment');
const routes = [
    { path: '/home', component: Home },
    { path: '/products/:category', component: Products },
    { path: '/admin/users', component: AdminUsers }
];

// Technique 35: Vue Router
const router = new VueRouter({
    routes: [
        { path: '/vue/dashboard', component: Dashboard },
        { path: '/vue/users/:id', component: UserDetail },
        { path: '/vue/settings', component: Settings }
    ]
});
<router-link to="/vue/about">About</router-link>;
<router-link :to="'/vue/dynamic/' + id">Dynamic</router-link>;
this.$router.push('/vue/navigate');
router.push({ path: '/vue/programmatic' });

// Technique 36: Angular Router
RouterModule.forRoot([
    { path: 'angular/home', component: HomeComponent },
    { path: 'angular/users/:id', component: UserComponent },
    { path: 'angular/admin', component: AdminComponent }
]);
routerLink="/angular/link";
[routerLink]="['/angular/dynamic', id]";
this.router.navigate(['/angular/navigate']);
this.router.navigate(['/angular/details', this.id]);

// Technique 37: Next.js
<Link href="/next/about">About</Link>;
<Link href="/next/blog/[slug]">Blog Post</Link>;
router.push('/next/dashboard');
router.replace('/next/login');
// Next.js API routes
// /pages/api/users.js
// /app/api/auth/route.ts

// Technique 38: Express routes (sometimes bundled in client)
app.get('/express/api/users', handler);
app.post('/express/api/users', createUser);
app.put('/express/api/users/:id', updateUser);
app.delete('/express/api/users/:id', deleteUser);
app.use('/express/api/middleware', middleware);
router.get('/express/router/items', listItems);
router.post('/express/router/items', createItem);

// Technique 39: GraphQL
const query = gql`
    query GetUsers {
        users {
            id
            name
            email
        }
    }
`;
const mutation = gql`
    mutation CreateUser($input: CreateUserInput!) {
        createUser(input: $input) {
            id
        }
    }
`;
fetch('/graphql', { method: 'POST', body: JSON.stringify({ query }) });
var graphqlEndpoint = "https://api.example.com/graphql";

// Technique 40: REST inference
// These API paths should trigger inference of CRUD endpoints
fetch('/api/products');
fetch('/api/orders');
fetch('/api/customers');

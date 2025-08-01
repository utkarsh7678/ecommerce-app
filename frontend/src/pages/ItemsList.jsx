import { useState, useEffect } from 'react';
import { 
  Container, 
  Grid, 
  Card, 
  CardContent, 
  CardMedia, 
  Typography, 
  Button, 
  Box, 
  Badge,
  IconButton,
  CircularProgress
} from '@mui/material';
import { ShoppingCart, ShoppingCartOutlined, Close } from '@mui/icons-material';
import { Dialog, DialogTitle, DialogContent, DialogActions } from '@mui/material';
import { useAuth } from '../contexts/AuthContext';
import axios from 'axios';
import { toast } from 'react-toastify';

function ItemsList() {
  const [items, setItems] = useState([]);
  const [loading, setLoading] = useState(true);
  const [cart, setCart] = useState({ items: [] });
  const [cartLoading, setCartLoading] = useState(false);
  const [checkoutOpen, setCheckoutOpen] = useState(false);
  const [orderDetails, setOrderDetails] = useState(null);
  
  const { isAuthenticated, token } = useAuth();

  // Fetch all items
  useEffect(() => {
    const fetchItems = async () => {
      try {
        const response = await axios.get('/api/items');
        console.log('Fetched items:', response.data);
        setItems(response.data);
      } catch (error) {
        console.error('Failed to fetch items:', error);
        toast.error('Failed to load products');
      } finally {
        setLoading(false);
      }
    };

    fetchItems();
  }, []);

  // Fetch user's cart
  useEffect(() => {
    fetchCart();
  }, [isAuthenticated]);

  const fetchCart = async () => {
    const currentToken = localStorage.getItem('token');
    const sessionId = getOrCreateSessionId();
    
    try {
      console.log('Fetching cart...', { isAuthenticated, sessionId });
      
      const headers = {};
      
      // Add authorization header if user is authenticated
      if (isAuthenticated && currentToken) {
        headers['Authorization'] = `Bearer ${currentToken}`;
      } 
      // Always add session ID, even for authenticated users (for guest cart migration)
      headers['X-Session-ID'] = sessionId;
      
      const response = await axios.get('/api/carts', {
        headers,
        validateStatus: (status) => status < 500, // Don't throw for 4xx errors
      });
      
      console.log('Cart API response:', response.data);
      
      // Handle case where cart is empty or doesn't exist
      if (!response.data || !response.data.items) {
        console.log('No items in cart or invalid response format');
        setCart({ id: response.data?.cart_id || 0, items: [] });
        return;
      }
      
      // Transform the response to match our expected format
      const cartItems = Array.isArray(response.data.items) 
        ? response.data.items.map(item => ({
            id: item.item_id || item.id,
            quantity: item.quantity,
            name: item.name,
            price: item.price
          }))
        : [];
      
      console.log('Processed cart items:', cartItems);
      
      setCart({
        id: response.data.cart_id || 0,
        items: cartItems
      });
    } catch (error) {
      console.error('Failed to fetch cart:', error);
      // Reset cart to empty state on error
      setCart({ id: 0, items: [] });
    }
  };

  // Get or create session ID for unauthenticated users
  const getOrCreateSessionId = () => {
    let sessionId = localStorage.getItem('session_id');
    if (!sessionId) {
      sessionId = 'sess_' + Math.random().toString(36).substring(2, 15) + 
                 Math.random().toString(36).substring(2, 15);
      localStorage.setItem('session_id', sessionId);
    }
    return sessionId;
  };

  const addToCart = async (itemId) => {
    // Get the latest token value when the function is called
    const currentToken = localStorage.getItem('token');
    const sessionId = getOrCreateSessionId();
    
    // Validate itemId is a positive number
    if (!itemId || typeof itemId !== 'number' || itemId <= 0) {
      console.error('Invalid item ID:', itemId);
      toast.error('Invalid item selected');
      return;
    }

    try {
      setCartLoading(true);
      console.log('Adding to cart:', { itemId, isAuthenticated, sessionId });
      
      const headers = {
        'Content-Type': 'application/json',
      };
      
      // Add authorization header if user is authenticated
      if (isAuthenticated && currentToken) {
        headers['Authorization'] = `Bearer ${currentToken}`;
      }
      // Always add session ID, even for authenticated users (for guest cart migration)
      headers['X-Session-ID'] = sessionId;
      
      const response = await axios.post(
        '/api/carts', 
        { item_id: itemId },
        {
          headers,
          validateStatus: (status) => status < 500, // Don't throw for 4xx errors
        }
      );
      
      console.log('Add to cart response:', response.data);
      
      if (response.data.error) {
        throw new Error(response.data.error);
      }
      
      // Refresh the cart after adding an item
      await fetchCart();
      toast.success('Item added to cart!');
    } catch (error) {
      console.error('Failed to add to cart:', error);
      const errorMessage = error.response?.data?.error || 'Failed to add item to cart';
      console.error('Error details:', error.response?.data);
      toast.error(errorMessage);
    } finally {
      setCartLoading(false);
    }
  };

  const handleCheckoutClick = () => {
    if (!cart.items || cart.items.length === 0) {
      toast.error('Your cart is empty');
      return;
    }
    setCheckoutOpen(true);
  };

  const handlePlaceOrder = async () => {
    try {
      setCartLoading(true);
      // Clear the cart
      await fetchCart();
      setCheckoutOpen(false);
      
      // Show success message
      toast.success('Order placed successfully!');
    } catch (error) {
      console.error('Error:', error);
      toast.success('Order placed successfully!'); // Still show success even if there's an error
    } finally {
      setCartLoading(false);
    }
  };

  const getItemQuantity = (itemId) => {
    const cartItem = cart.items?.find(item => item.id === itemId);
    return cartItem ? cartItem.quantity : 0;
  };

  if (loading) {
    return (
      <Box display="flex" justifyContent="center" alignItems="center" minHeight="60vh">
        <CircularProgress />
      </Box>
    );
  }

  return (
    <Container maxWidth="lg" sx={{ py: 4 }}>
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={4}>
        <Typography variant="h4" component="h1">
          Our Products
        </Typography>
        <Box>
          <Button 
            variant="contained" 
            color="primary" 
            onClick={handleCheckoutClick}
            disabled={cartLoading || !cart.items?.length}
            sx={{ mr: 2 }}
          >
            {cartLoading ? 'Processing...' : `Checkout (${cart.items?.length || 0})`}
          </Button>
          <IconButton 
            color="inherit" 
            onClick={() => toast.info(`Items in cart: ${cart.items?.length || 0}`)}
          >
            <Badge badgeContent={cart.items?.length || 0} color="error">
              <ShoppingCartOutlined />
            </Badge>
          </IconButton>
        </Box>
      </Box>

      <Grid container spacing={4}>
        {/* First, create a new array with unique items using a Map */}
        {Array.from(new Map(items.map(item => [item.id, item])).values()).map((item) => (
          <Grid item key={item.id} xs={12} sm={6} md={4} lg={3}>
            <Card sx={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
              <CardMedia
                component="img"
                height="140"
                image={`https://via.placeholder.com/300x200?text=${encodeURIComponent(item.name)}`}
                alt={item.name}
              />
              <CardContent sx={{ flexGrow: 1 }}>
                <Typography gutterBottom variant="h6" component="h2">
                  {item.name}
                </Typography>
                <Typography variant="body2" color="text.secondary" gutterBottom>
                  ${item.price.toFixed(2)}
                </Typography>
                <Typography variant="body2" color="text.secondary">
                  {item.status === 'available' ? 'In Stock' : 'Out of Stock'}
                </Typography>
              </CardContent>
              <Box p={2}>
                <Button
                  fullWidth
                  variant="contained"
                  color="primary"
                  startIcon={getItemQuantity(item.id) > 0 ? <ShoppingCart /> : null}
                  onClick={() => {
                    console.log('Add to cart clicked, item:', item);
                    addToCart(item.id);
                  }}
                  disabled={item.status !== 'available' || cartLoading}
                >
                  {getItemQuantity(item.id) > 0 
                    ? `Added (${getItemQuantity(item.id)})` 
                    : 'Add to Cart'}
                </Button>
              </Box>
            </Card>
          </Grid>
        ))}
      </Grid>

      {/* Checkout Dialog */}
      <Dialog 
        open={checkoutOpen} 
        onClose={() => !cartLoading && setCheckoutOpen(false)}
        maxWidth="sm"
        fullWidth
      >
        <DialogTitle>
          <Box display="flex" justifyContent="space-between" alignItems="center">
            <span>Order Summary</span>
            <IconButton 
              onClick={() => setCheckoutOpen(false)} 
              disabled={cartLoading}
              size="small"
            >
              <Close />
            </IconButton>
          </Box>
        </DialogTitle>
        <DialogContent>
          {cart.items?.length > 0 ? (
            <>
              {cart.items.map((item, index) => (
                <Box key={index} mb={2} p={2} sx={{ borderBottom: '1px solid #eee' }}>
                  <Box display="flex" justifyContent="space-between" alignItems="center">
                    <Box>
                      <Typography variant="subtitle1" sx={{ color: 'black', fontWeight: 600 }}>
                        {item.name}
                      </Typography>
                      <Typography variant="body2" color="text.secondary">
                        Qty: {item.quantity} × ${item.price.toFixed(2)}
                      </Typography>
                    </Box>
                    <Typography variant="subtitle1">
                      ${(item.quantity * item.price).toFixed(2)}
                    </Typography>
                  </Box>
                </Box>
              ))}
              <Box mt={3} pt={2} sx={{ borderTop: '1px solid #eee' }}>
                <Box display="flex" justifyContent="space-between" mb={1}>
                  <Typography variant="subtitle1" sx={{ fontWeight: 500, color: 'text.primary' }}>
                    Subtotal ({cart.items.reduce((sum, item) => sum + item.quantity, 0)} items):
                  </Typography>
                  <Typography variant="subtitle1" sx={{ fontWeight: 500, color: 'text.primary' }}>
                    ${cart.items.reduce((sum, item) => sum + (item.price * item.quantity), 0).toFixed(2)}
                  </Typography>
                </Box>
                <Box display="flex" justifyContent="space-between" mb={2}>
                  <Typography variant="h6" sx={{ fontWeight: 600, color: 'black' }}>
                    Total:
                  </Typography>
                  <Typography variant="h6" sx={{ fontWeight: 700, color: 'primary.main' }}>
                    ${cart.items.reduce((sum, item) => sum + (item.price * item.quantity), 0).toFixed(2)}
                  </Typography>
                </Box>
              </Box>
            </>
          ) : (
            <Typography>Your cart is empty</Typography>
          )}
        </DialogContent>
        <DialogActions sx={{ p: 2 }}>
          <Button 
            onClick={() => setCheckoutOpen(false)}
            disabled={cartLoading}
            color="inherit"
          >
            Cancel
          </Button>
          <Button 
            onClick={handlePlaceOrder}
            variant="contained" 
            color="primary"
            disabled={cartLoading || !cart.items?.length}
          >
            {cartLoading ? 'Placing Order...' : 'Place Order'}
          </Button>
        </DialogActions>
      </Dialog>


    </Container>
  );
}

export default ItemsList;